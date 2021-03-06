package cross_set

import (
	"github.com/rs/zerolog/log"

	"github.com/domino14/cwgame/alphabet"
	"github.com/domino14/cwgame/board"
	"github.com/domino14/cwgame/move"
)

type CrossSet = board.CrossSet
type Board = board.GameBoard

const (
	Left       = board.LeftDirection
	Right      = board.RightDirection
	Horizontal = board.HorizontalDirection
	Vertical   = board.VerticalDirection
)

// Public cross_set.Generator Interface
// There are two concrete implementations below,
// - CrossScoreOnlyGenerator{Dist}
// - GaddagCrossSetGenerator{Dist, Gaddag}

type Generator interface {
	Generate(b *Board, row int, col int, dir board.BoardDirection)
	GenerateAll(b *Board)
	UpdateForMove(b *Board, m *move.Move)
}

// We have to go through this dance since go will not let us simply provide
// Generator with default implementations of GenerateAll and UpdateForMove that
// call a given implementation of Generate.

type iGenerator interface {
	Generate(b *Board, row int, col int, dir board.BoardDirection)
}

// generateAll generates all cross-sets. It goes through the entire
// board for both transpositions.
func generateAll(g iGenerator, b *Board) {
	n := b.Dim()
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			g.Generate(b, i, j, Horizontal)
		}
	}
	b.Transpose()
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			g.Generate(b, i, j, Vertical)
		}
	}
	// And transpose back to the original orientation.
	b.Transpose()
}

func updateForMove(g iGenerator, b *Board, m *move.Move) {

	log.Debug().Msgf("Updating for move: %s", m.ShortDescription())
	row, col, vertical := m.CoordsAndVertical()
	// Every tile placed by this new move creates new "across" words, and we need
	// to update the cross sets on both sides of these across words, as well
	// as the cross sets for THIS word.

	// Assumes all across words are HORIZONTAL.
	calcForAcross := func(rowStart int, colStart int, csd board.BoardDirection) {
		for row := rowStart; row < len(m.Tiles())+rowStart; row++ {
			if m.Tiles()[row-rowStart] == alphabet.PlayedThroughMarker {
				// No new "across word" was generated by this tile, so no need
				// to update cross set.
				continue
			}
			// Otherwise, look along this row. Note, the edge is still part
			// of the word.
			rightCol := b.WordEdge(int(row), int(colStart), Right)
			leftCol := b.WordEdge(int(row), int(colStart), Left)
			g.Generate(b, int(row), int(rightCol)+1, csd)
			g.Generate(b, int(row), int(leftCol)-1, csd)
			// This should clear the cross set on the just played tile.
			g.Generate(b, int(row), int(colStart), csd)
		}
	}

	// assumes self is HORIZONTAL
	calcForSelf := func(rowStart int, colStart int, csd board.BoardDirection) {
		// Generate cross-sets on either side of the word.
		for col := int(colStart) - 1; col <= int(colStart)+len(m.Tiles()); col++ {
			g.Generate(b, int(rowStart), col, csd)
		}
	}

	if vertical {
		calcForAcross(row, col, Horizontal)
		b.Transpose()
		row, col = col, row
		calcForSelf(row, col, Vertical)
		b.Transpose()
	} else {
		calcForSelf(row, col, Horizontal)
		b.Transpose()
		row, col = col, row
		calcForAcross(row, col, Vertical)
		b.Transpose()
	}
}

// ----------------------------------------------------------------------
// Use a CrossScoreOnlyGenerator when you don't need cross sets

type CrossScoreOnlyGenerator struct {
	Dist *alphabet.LetterDistribution
}

func (g CrossScoreOnlyGenerator) Generate(b *Board, row int, col int, dir board.BoardDirection) {
	genCrossScore(b, row, col, dir, g.Dist)
}

func (g CrossScoreOnlyGenerator) GenerateAll(b *Board) {
	generateAll(g, b)
}

func (g CrossScoreOnlyGenerator) UpdateForMove(b *Board, m *move.Move) {
	updateForMove(g, b, m)
}

// Wrapper functions to save rewriting all the tests

func GenAllCrossScores(b *Board, ld *alphabet.LetterDistribution) {
	gen := CrossScoreOnlyGenerator{Dist: ld}
	gen.GenerateAll(b)
}

// ----------------------------------------------------------------------
// Implementation for CrossScoreOnlyGenerator

func genCrossScore(b *Board, row int, col int, dir board.BoardDirection,
	ld *alphabet.LetterDistribution) {
	if row < 0 || row >= b.Dim() || col < 0 || col >= b.Dim() {
		return
	}
	// If the square has a letter in it, its cross set and cross score
	// should both be 0
	if !b.GetSquare(row, col).IsEmpty() {
		b.GetSquare(row, col).SetCrossScore(0, dir)
		return
	}
	// If there's no tile adjacent to this square in any direction,
	// every letter is allowed.
	if b.LeftAndRightEmpty(row, col) {
		b.GetSquare(row, col).SetCrossScore(0, dir)
		return
	}
	// If we are here, there is a letter to the left, to the right, or both.
	// start from the right and go backwards.
	rightCol := b.WordEdge(row, col+1, Right)
	if rightCol == col {
		score := b.TraverseBackwardsForScore(row, col-1, ld)
		b.GetSquare(row, col).SetCrossScore(score, dir)
	} else {
		// Otherwise, the right is not empty. Check if the left is empty,
		// if so we just traverse right, otherwise, we try every letter.
		scoreR := b.TraverseBackwardsForScore(row, rightCol, ld)
		scoreL := b.TraverseBackwardsForScore(row, col-1, ld)
		b.GetSquare(row, col).SetCrossScore(scoreR+scoreL, dir)
	}
}
