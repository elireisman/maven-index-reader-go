package resources

import "io"

// Resource -
type Resource interface {
	// Obtain a Reader on the given Resource, or an error
	Read() (io.ReadSeekCloser, error)

	// Close the resource
	Close() error
}
