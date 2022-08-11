package readers

import (
	"compress/gzip"
	"io"
	"log"
	"time"

	"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type ChunkReader struct {
	logger   *log.Logger
	resource resources.Resource
	buffer   chan data.Record
}

// NewChunkReader -
func NewChunkReader(l *log.Logger, r resources.Resource, b chan data.Record) ChunkReader {
	return ChunkReader{
		logger:   l,
		resource: r,
		buffer:   b,
	}
}

// Get - obtain read-only channel the caller can consume data.Records from
func (cr ChunkReader) Get() <-chan data.Record {
	return cr.buffer
}

// Read - initiate async consumption of Resource and population of data.Record buffer
func (cr ChunkReader) Read() error {
	if cr.resource == nil {
		return errors.New("ChunkReader: cannot read from nil Resource")
	}

	rdr, err := cr.resource.Reader()
	if err != nil {
		return errors.Wrapf(err, "ChunkReader: failed to obtain data stream from %s with cause", cr.resource)
	}

	gzRdr, err := gzip.NewReader(rdr)
	if err != nil {
		cr.resource.Close()
		return errors.Wrapf(err, "ChunkReader: failed to wrap %s in GZIP Reader with cause", cr.resource)
	}

	var chunkVersion uint8
	if b, err := utils.ReadByte(gzRdr); err == nil {
		chunkVersion = uint8(b)
	} else {
		gzRdr.Close()
		return errors.Wrap(err, "ChunkReader: failed to read chunk version with cause")
	}

	var chunkTimestamp time.Time
	if i64, err := utils.ReadInt64(gzRdr); err == nil {
		secs := i64 / 1000
		nanos := (i64 % 1000) * 1000000
		chunkTimestamp = time.Unix(secs, nanos)
	} else {
		gzRdr.Close()
		return errors.Wrap(err, "ChunkReader: failed to read chunk timestamp with cause")
	}

	// this goroutine now owns GZIP Reader and must close it
	cr.logger.Printf("ChunkReader: found %s of version %d at time %s", cr.resource, chunkVersion, chunkTimestamp)
	go cr.recordIterator(gzRdr, chunkVersion, chunkTimestamp)

	return nil
}

func (cr ChunkReader) recordIterator(gzRdr io.ReadCloser, chunkVersion uint8, chunkTimestamp time.Time) {
	defer func() {
		gzRdr.Close()
		close(cr.buffer)
	}()

	var err error
	count := 1
	for {
		var fieldCount int32
		fieldCount, err = utils.ReadInt32(gzRdr)
		if err != nil {
			cr.logger.Panicf(
				"ChunkReader: failed to read field count for record %d from %s with cause: %s",
				count, cr.resource, err)
		}

		rawRecord := map[string]string{}
		for ndx := int32(0); ndx < fieldCount; ndx++ {
			_, err = utils.ReadByte(gzRdr)
			if err != nil {
				cr.logger.Panicf(
					"ChunkReader: failed to read field flags for record %d from %s with cause: %s",
					count, cr.resource, err)
			}

			var key string
			key, err = utils.ReadString(gzRdr)
			if err != nil {
				cr.logger.Panicf(
					"ChunkReader: failed to read field key for record %d from %s with cause: %s",
					count, cr.resource, err)
			}

			var value string
			value, err = utils.ReadString(gzRdr)
			if err != nil && err != io.EOF {
				cr.logger.Panicf(
					"ChunkReader: failed to read field value for key %s on record %d from %s with cause: %s",
					key, count, cr.resource, err)
			}

			rawRecord[key] = value
		}

		record, rErr := data.NewRecord(rawRecord)
		if rErr != nil {
			cr.logger.Panicf(
				"ChunkReader: failed to compose well-formed record %d from %s from %s with cause: %s (at EOF: %t)",
				count, rawRecord, cr.resource, rErr, err == io.EOF)
		}
		cr.buffer <- record

		if err == io.EOF {
			cr.logger.Printf("ChunkReader: successfully published %d records from %s", count, cr.resource)
			return
		}
		count++
	}
}
