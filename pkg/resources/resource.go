package resources

import (
	"io"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"

	"github.com/pkg/errors"
)

// Resource -
type Resource interface {
	// Obtain a Reader on the given Resource, or an error
	Reader() (io.Reader, error)

	// Close the resource
	Close() error
}

// resolve a Resource from caller-supplied config.Index
func FromConfig(logger *log.Logger, cfg config.Index, target string) (Resource, error) {
	var resource Resource
	var err error

	switch cfg.Source.Type {
	case config.Local:
		resource, err = NewLocalResource(logger, target)
	case config.HTTP:
		resource, err = NewHttpResource(logger, target)
	default:
		err = errors.Errorf("ConfigureResource: invalid config.Index.Source.Type for target %s, got: %d", target, cfg.Source.Type)
	}

	return resource, err
}
