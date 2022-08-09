package resources

import (
	"github.com/elireisman/maven-index-reader-go/internal/reader"
)

// Resource -
type Resource interface {
	// Obtain a Reader on the given Resource, or an error
	Reader() (reader.Maven, error)

	// Close the resource
	Close() error
}
