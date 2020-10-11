package game

import (
	"strings"

	"github.com/domino14/cwgame/board"
	pb "github.com/domino14/cwgame/gen/proto/cwgame"
)

// HistoryToVariant takes in a game history and returns the board configuration
// and letter distribution name.
func HistoryToVariant(h *pb.GameHistory) (boardLayout []string, letterDistributionName string) {

	switch h.Variant {
	case "CrosswordGame":
		boardLayout = board.CrosswordGameBoard
	default:
		boardLayout = board.CrosswordGameBoard
	}
	letterDistributionName = "english"
	switch {
	case strings.HasPrefix(h.Lexicon, "OSPS"):
		letterDistributionName = "polish"
	case strings.HasPrefix(h.Lexicon, "FISE"):
		letterDistributionName = "spanish"
	}
	return boardLayout, letterDistributionName
}
