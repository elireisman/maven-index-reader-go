package output

import (
	"log"
	"strings"

	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

// Format - contract for supported ouput formats
type Format interface {
	Write() error
}

// ResolveFormat -
func ResolveFormat(logger *log.Logger, queue <-chan data.Record, specifier, filePath string) (Format, error) {
	switch strings.ToLower(specifier) {
	case "json":
		return NewJSON(logger, queue, filePath), nil
	case "csv":
		return NewCSV(logger, queue, filePath), nil
	case "log":
		return NewLogger(logger, queue), nil
	default:
		// fall through
	}

	return nil, errors.Errorf("invalid output format: %s", specifier)
}
