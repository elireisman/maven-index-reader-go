package output

import (
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
)

// Format - contract for supported ouput formats
type Format interface {
	Write() error
}

// ResolveFormat -
func ResolveFormat(logger *log.Logger, queue <-chan data.Record, cfg config.Index) Format {
	var out Format

	switch cfg.Output.Format {
	case config.JSON:
		out = NewJSON(logger, queue, cfg)
	case config.CSV:
		out = NewCSV(logger, queue, cfg)
	default: // log unformatted Go structs
		out = NewLogger(logger, queue, cfg)
	}

	return out
}
