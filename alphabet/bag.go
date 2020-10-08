package alphabet

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/rs/zerolog/log"
)

// A Bag is the bag o'tiles!
type Bag struct {
	initialTiles       []MachineLetter
	tiles              []MachineLetter
	initialTileMap     map[MachineLetter]uint8
	tileMap            map[MachineLetter]uint8
	letterDistribution *LetterDistribution
	randSource         *rand.Rand
}

func copyTileMap(orig map[MachineLetter]uint8) map[MachineLetter]uint8 {
	tm := make(map[MachineLetter]uint8)
	for k, v := range orig {
		tm[k] = v
	}
	return tm
}

// Refill refills the bag.
func (b *Bag) Refill() {
	b.tiles = append([]MachineLetter(nil), b.initialTiles...)
	b.tileMap = copyTileMap(b.initialTileMap)
	b.Shuffle()
}

// DrawAtMost draws at most n tiles from the bag. It can draw fewer if there
// are fewer tiles than n, and even draw no tiles at all :o
func (b *Bag) DrawAtMost(n int) []MachineLetter {
	if n > len(b.tiles) {
		n = len(b.tiles)
	}
	drawn, _ := b.Draw(n)
	return drawn
}

// Draw draws n tiles from the bag.
func (b *Bag) Draw(n int) ([]MachineLetter, error) {
	if n > len(b.tiles) {
		return nil, fmt.Errorf("tried to draw %v tiles, tile bag has %v",
			n, len(b.tiles))
	}
	drawn := make([]MachineLetter, n)
	for i := 0; i < n; i++ {
		drawn[i] = b.tiles[i]
		b.tileMap[drawn[i]]--
	}
	b.tiles = b.tiles[n:]
	// log.Debug().Int("numtiles", len(b.tiles)).Int("drew", n).Msg("drew from bag")
	return drawn, nil
}

func (b *Bag) Peek() []MachineLetter {
	ret := make([]MachineLetter, len(b.tiles))
	copy(ret, b.tiles)
	return ret
}

// Shuffle shuffles the bag.
func (b *Bag) Shuffle() {
	// log.Debug().Int("numtiles", len(b.tiles)).Msg("shuffling bag")
	b.randSource.Shuffle(len(b.tiles), func(i, j int) {
		b.tiles[i], b.tiles[j] = b.tiles[j], b.tiles[i]
	})
}

// Exchange exchanges the junk in your rack with new tiles.
func (b *Bag) Exchange(letters []MachineLetter) ([]MachineLetter, error) {
	newTiles, err := b.Draw(len(letters))
	if err != nil {
		return nil, err
	}
	// put exchanged tiles back into the bag and re-shuffle
	b.PutBack(letters)
	return newTiles, nil
}

// PutBack puts the tiles back in the bag, and shuffles the bag.
func (b *Bag) PutBack(letters []MachineLetter) {
	if len(letters) == 0 {
		return
	}
	b.tiles = append(b.tiles, letters...)
	for _, ml := range letters {
		b.tileMap[ml]++
	}
	b.Shuffle()
}

// hasRack returns a boolean indicating whether the passed-in rack is
// in the bag, in its entirety.
func (b *Bag) hasRack(letters []MachineLetter) bool {
	submap := make(map[MachineLetter]uint8)

	for _, ml := range letters {
		if ml.IsBlanked() {
			submap[BlankMachineLetter]++
		} else {
			submap[ml]++
		}
	}
	// check every single letter we have.
	for ml, ct := range submap {
		if b.tileMap[ml] < ct {
			return false
		}
	}

	return true
}

func (b *Bag) TilesRemaining() int {
	return len(b.tiles)
}

func (b *Bag) remove(t MachineLetter) {
	if b.tileMap[t] == 0 {
		log.Fatal().Msgf("Tile %c not found in bag", t)
	}
	b.tileMap[t]--
}

// rebuildTileSlice reconciles the bag slice with the tile map.
func (b *Bag) rebuildTileSlice(numTilesInBag int) error {
	log.Debug().Msgf("reconciling tiles, num in bag are %v, map %v",
		numTilesInBag, b.tileMap)
	if numTilesInBag > len(b.initialTiles) {
		return errors.New("more tiles in the bag that there were to begin with")
	}
	b.tiles = make([]MachineLetter, numTilesInBag)
	idx := 0
	for let, ct := range b.tileMap {
		for j := uint8(0); j < ct; j++ {
			b.tiles[idx] = let
			idx++
		}
	}
	b.Shuffle()
	return nil
}

// Redraw is basically a do-over; throw the current rack in the bag
// and draw a new rack.
func (b *Bag) Redraw(currentRack []MachineLetter) []MachineLetter {
	b.PutBack(currentRack)
	return b.DrawAtMost(7)
}

// RemoveTiles removes the given tiles from the bag, and returns an error
// if it can't.
func (b *Bag) RemoveTiles(tiles []MachineLetter) error {
	if !b.hasRack(tiles) {
		return fmt.Errorf("cannot remove the tiles %v from the bag, as they are not in the bag",
			MachineWord(tiles).UserVisible(b.LetterDistribution().alph))
	}
	for _, t := range tiles {
		if t.IsBlanked() {
			b.remove(BlankMachineLetter)
		} else {
			b.remove(t)
		}
	}
	return b.rebuildTileSlice(len(b.tiles) - len(tiles))
}

func NewBag(ld *LetterDistribution, alph *Alphabet, randSource *rand.Rand) *Bag {

	tiles := make([]MachineLetter, ld.numLetters)
	tileMap := map[MachineLetter]uint8{}

	idx := 0
	for rn, ct := range ld.Distribution {
		val, err := alph.Val(rn)
		if err != nil {
			log.Fatal().Msgf("Attempt to initialize bag failed: %v", err)
		}
		tileMap[val] = ct
		for j := uint8(0); j < ct; j++ {
			tiles[idx] = val
			idx++
		}
	}

	return &Bag{
		tiles:              tiles,
		tileMap:            tileMap,
		initialTiles:       append([]MachineLetter(nil), tiles...),
		initialTileMap:     copyTileMap(tileMap),
		letterDistribution: ld,
		randSource:         randSource,
	}
}

// Copy copies to a new bag and returns it. Note that the initialTiles
// are only shallowly copied. This is fine because
// we don't ever expect these to change after initialization.
// If randSource is not nil, it is set as the rand source for the copy.
// Otherwise, use the original's rand source.
func (b *Bag) Copy(randSource *rand.Rand) *Bag {
	tiles := make([]MachineLetter, len(b.tiles))
	tileMap := make(map[MachineLetter]uint8)
	copy(tiles, b.tiles)
	// Copy map as well
	for k, v := range b.tileMap {
		tileMap[k] = v
	}
	if randSource == nil {
		randSource = b.randSource
	}

	return &Bag{
		tiles:              tiles,
		tileMap:            tileMap,
		initialTiles:       b.initialTiles,
		initialTileMap:     b.initialTileMap,
		letterDistribution: b.letterDistribution,
		randSource:         randSource,
	}
}

// CopyFrom copies back the tiles from another bag into this bag. The caller
// of this function is responsible for ensuring `other` has the other
// structures we need! (letter distribution, etc).
// It should have been created from the Copy function above.
func (b *Bag) CopyFrom(other *Bag) {
	// This is a deep copy and can be kind of wasteful, but we don't use
	// the bag often.
	if len(other.tiles) == 0 {
		b.tiles = []MachineLetter{}
		b.tileMap = map[MachineLetter]uint8{}
		return
	}
	b.tiles = make([]MachineLetter, len(other.tiles))
	copy(b.tiles, other.tiles)
	b.tileMap = copyTileMap(other.tileMap)
	b.randSource = other.randSource
}

func (b *Bag) LetterDistribution() *LetterDistribution {
	return b.letterDistribution
}
