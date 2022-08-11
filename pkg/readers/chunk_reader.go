package readers

import (
	"log"

	//"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type ChunkReader struct {
	logger   *log.Logger
	resource resources.Resource
}

func NewChunkReader(l *log.Logger, r resources.Resource) ChunkReader {
	return ChunkReader{
		logger:   l,
		resource: r,
	}
}

// TODO(eli): IMPLMENT THIS
func (cr ChunkReader) Read() (<-chan data.Record, error) {
	return nil, errors.New("not implemented")
}
