package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Debug                     bool
	LetterDistributionPath    string
	DefaultLetterDistribution string
	DefaultLexicon            string
}

// Default config from environment variables. Since the config struct is
// mutable we don't just make this a global shared variable, but provide a
// factory function to copy it on every call.
var defaultConfig = Config{
	Debug:                     false,
	LetterDistributionPath:    os.Getenv("LETTER_DISTRIBUTION_PATH"),
	DefaultLetterDistribution: "English",
	DefaultLexicon:            "NWL18",
}

func DefaultConfig() Config {
	return defaultConfig
}

func (c *Config) AdjustRelativePaths(basepath string) {
	basepath = FindBasePath(basepath)
	c.LetterDistributionPath = toAbsPath(basepath, c.LetterDistributionPath, "ldpath")
}

func FindBasePath(path string) string {
	// Search up a path until we find the toplevel dir with data/ under it.
	// This will likely do bad things if there is no such dir but right now we
	// are running stuff from within the cwgame directory and ultimately we want
	// to use something like $HOME/.cwgame anyway rather than the exe path.
	for {
		data := filepath.Join(path, "data")
		_, err := os.Stat(data)
		if !(os.IsNotExist(err)) {
			break
		}
		path = filepath.Dir(path)
	}
	return path
}

func toAbsPath(basepath string, path string, logname string) string {
	if strings.HasPrefix(path, "./") {
		path = filepath.Join(basepath, path)
		log.Info().Str(logname, path).Msgf("adjusted relative path")
	}
	return path
}
