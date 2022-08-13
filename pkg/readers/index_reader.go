package readers

import (
	"log"

	//"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type IndexReader struct {
	cfg      config.Index
	logger   *log.Logger
	resource resources.Resource
	buffer   chan data.Chunk
}

func NewIndexReader(l *log.Logger, r resources.Resource, b chan data.Chunk, c config.Index) IndexReader {
	return IndexReader{
		cfg:      c,
		logger:   l,
		resource: r,
		buffer:   b,
	}
}

func (ir IndexReader) Get() <-chan data.Chunk {
	return ir.buffer
}

// TODO(eli): IMPLMENT THIS
func (ir IndexReader) Read() error {
	return errors.New("not implemented")
}
