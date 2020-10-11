package cross_set

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/domino14/cwgame/alphabet"
	"github.com/domino14/cwgame/board"
	"github.com/domino14/cwgame/config"
	"github.com/domino14/cwgame/move"
)

var DefaultConfig = config.DefaultConfig()

const (
	VsEd     = board.VsEd
	VsJeremy = board.VsJeremy
	VsMatt   = board.VsMatt
	VsMatt2  = board.VsMatt2
	VsOxy    = board.VsOxy
)

type crossSetTestCase struct {
	row      int
	col      int
	crossSet board.CrossSet
	dir      board.BoardDirection
	score    int
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGenAllCrossScores(t *testing.T) {
	dist, err := alphabet.EnglishLetterDistribution(&DefaultConfig)
	if err != nil {
		t.Error(err)
	}
	alph := dist.Alphabet()

	b := board.MakeBoard(board.CrosswordGameBoard)
	b.SetToGame(alph, VsEd)

	GenAllCrossScores(b, dist)

	var testCases = []crossSetTestCase{
		{8, 8, board.CrossSetFromString("OS", alph), board.HorizontalDirection, 8},
		{8, 8, board.CrossSetFromString("S", alph), board.VerticalDirection, 9},
		{5, 11, board.CrossSetFromString("S", alph), board.HorizontalDirection, 5},
		{5, 11, board.CrossSetFromString("AO", alph), board.VerticalDirection, 2},
		{8, 13, board.CrossSetFromString("AEOU", alph), board.HorizontalDirection, 1},
		{8, 13, board.CrossSetFromString("AEIMOUY", alph), board.VerticalDirection, 3},
		{9, 13, board.CrossSetFromString("HMNPST", alph), board.HorizontalDirection, 1},
		{9, 13, board.TrivialCrossSet, board.VerticalDirection, 0},
		{14, 14, board.TrivialCrossSet, board.HorizontalDirection, 0},
		{14, 14, board.TrivialCrossSet, board.VerticalDirection, 0},
		{12, 12, board.CrossSet(0), board.HorizontalDirection, 0},
		{12, 12, board.CrossSet(0), board.VerticalDirection, 0},
	}

	for _, tc := range testCases {
		// Compare values
		if b.GetCrossScore(tc.row, tc.col, tc.dir) != tc.score {
			t.Errorf("For row=%v col=%v, Expected cross-score to be %v, got %v",
				tc.row, tc.col, tc.score,
				b.GetCrossScore(tc.row, tc.col, tc.dir))
		}
	}
	// This one has more nondeterministic (in-between LR) crosssets
	b.SetToGame(alph, VsMatt)
	GenAllCrossScores(b, dist)
	testCases = []crossSetTestCase{
		{8, 7, board.CrossSetFromString("S", alph), board.HorizontalDirection, 11},
		{8, 7, board.CrossSet(0), board.VerticalDirection, 12},
		{5, 11, board.CrossSetFromString("BGOPRTWX", alph), board.HorizontalDirection, 2},
		{5, 11, board.CrossSet(0), board.VerticalDirection, 15},
		{8, 13, board.TrivialCrossSet, board.HorizontalDirection, 0},
		{8, 13, board.TrivialCrossSet, board.VerticalDirection, 0},
		{11, 4, board.CrossSetFromString("DRS", alph), board.HorizontalDirection, 6},
		{11, 4, board.CrossSetFromString("CGM", alph), board.VerticalDirection, 1},
		{2, 2, board.TrivialCrossSet, board.HorizontalDirection, 0},
		{2, 2, board.CrossSetFromString("AEI", alph), board.VerticalDirection, 2},
		{7, 12, board.CrossSetFromString("AEIOY", alph), board.HorizontalDirection, 0}, // it's a blank
		{7, 12, board.TrivialCrossSet, board.VerticalDirection, 0},
		{11, 8, board.CrossSet(0), board.HorizontalDirection, 4},
		{11, 8, board.CrossSetFromString("AEOU", alph), board.VerticalDirection, 1},
		{1, 8, board.CrossSetFromString("AEO", alph), board.HorizontalDirection, 1},
		{1, 8, board.CrossSetFromString("DFHLMNRSTX", alph), board.VerticalDirection, 1},
		{10, 10, board.CrossSetFromString("E", alph), board.HorizontalDirection, 11},
		{10, 10, board.TrivialCrossSet, board.VerticalDirection, 0},
	}
	for _, tc := range testCases {
		if b.GetCrossScore(tc.row, tc.col, tc.dir) != tc.score {
			t.Errorf("For row=%v col=%v, Expected cross-score to be %v, got %v",
				tc.row, tc.col, tc.score,
				b.GetCrossScore(tc.row, tc.col, tc.dir))
		}
	}
}

type updateCrossesForMoveTestCase struct {
	testGame        board.VsWho
	m               *move.Move
	userVisibleWord string
}

func TestUpdateCrossScoresForMove(t *testing.T) {
	dist, err := alphabet.EnglishLetterDistribution(&DefaultConfig)
	if err != nil {
		t.Error(err)
	}
	gen := CrossScoreOnlyGenerator{Dist: dist}
	alph := dist.Alphabet()

	var testCases = []updateCrossesForMoveTestCase{
		{VsMatt, move.NewScoringMoveSimple(38, "K9", "TAEL", "ABD", alph), "TAEL"},
		// Test right edge of board
		{VsMatt2, move.NewScoringMoveSimple(77, "O8", "TENsILE", "", alph), "TENsILE"},
		// Test through tiles
		{VsOxy, move.NewScoringMoveSimple(1780, "A1", "OX.P...B..AZ..E", "", alph),
			"OXYPHENBUTAZONE"},
		// Test top of board, horizontal
		{VsJeremy, move.NewScoringMoveSimple(14, "1G", "S.oWED", "D?", alph), "SNoWED"},
		// Test bottom of board, horizontal
		{VsJeremy, move.NewScoringMoveSimple(11, "15F", "F..ER", "", alph), "FOYER"},
	}

	// create a move.
	for _, tc := range testCases {
		b := board.MakeBoard(board.CrosswordGameBoard)
		b.SetToGame(alph, tc.testGame)
		gen.GenerateAll(b)
		b.UpdateAllAnchors()
		b.PlayMove(tc.m, dist)
		gen.UpdateForMove(b, tc.m)
		log.Printf(b.ToDisplayText(alph))
		// Create an identical board, but generate cross-sets for the entire
		// board after placing the letters "manually".
		c := board.MakeBoard(board.CrosswordGameBoard)
		c.SetToGame(alph, tc.testGame)
		c.PlaceMoveTiles(tc.m)
		c.TestSetTilesPlayed(c.GetTilesPlayed() + tc.m.TilesPlayed())
		GenAllCrossScores(c, dist)
		c.UpdateAllAnchors()

		assert.True(t, b.Equals(c))

		for i, c := range tc.userVisibleWord {
			row, col, vertical := tc.m.CoordsAndVertical()
			var rowInc, colInc int
			if vertical {
				rowInc = i
				colInc = 0
			} else {
				rowInc = 0
				colInc = i
			}
			uv := b.GetSquare(row+rowInc, col+colInc).Letter().UserVisible(alph)
			assert.Equal(t, c, uv)
		}
	}
}
