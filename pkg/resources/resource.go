package resources

import (
	"io"
)

// Resource -
type Resource interface {
	// Obtain a Reader on the given Resource, or an error
	Reader() (io.Reader, error)

	// Close the resource
	Close() error
}
