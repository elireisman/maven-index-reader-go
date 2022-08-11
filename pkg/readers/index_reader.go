package readers

import (
	"log"

	//"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type IndexReader struct {
	logger   *log.Logger
	resource resources.Resource
}

func NewIndexReader(l *log.Logger, r resources.Resource) IndexReader {
	return IndexReader{
		logger:   l,
		resource: r,
	}
}

// TODO(eli): IMPLMENT THIS
func (ir IndexReader) Read() (<-chan data.Chunk, error) {
	return nil, errors.New("not implemented")
}
