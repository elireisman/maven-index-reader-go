package readers

import (
	"compress/gzip"
	"io"
	"log"
	"time"

	"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type Chunk struct {
	target   string
	cfg      config.Index
	logger   *log.Logger
	buffer   chan<- data.Record
	filterFn FilterFunc
}

// caller-defined filter on data.Records extracted
// from the chunk. Returning false drops the record
type FilterFunc func(data.Record) bool

// NewChunk - caller supplies the input resource as well as the
// output channel for captured records that the caller plans to consume
func NewChunk(l *log.Logger, b chan<- data.Record, c config.Index, t string, ff FilterFunc) Chunk {
	return Chunk{
		target:   t,
		cfg:      c,
		logger:   l,
		buffer:   b,
		filterFn: ff,
	}
}

// Read - initiate async consumption of Resource and population of data.Record buffer
func (cr Chunk) Read() error {
	resource, err := resources.FromConfig(cr.logger, cr.cfg, cr.target)
	defer resource.Close()

	rdr, err := resource.Reader()
	if err != nil {
		return errors.Wrapf(err, "Chunk: failed to obtain data stream from %s with cause", resource)
	}

	gzRdr, err := gzip.NewReader(rdr)
	if err != nil {
		return errors.Wrapf(err, "Chunk: failed to wrap %s in GZIP Reader with cause", resource)
	}
	defer gzRdr.Close()

	var chunkVersion uint8
	if b, err := utils.ReadByte(gzRdr); err == nil {
		chunkVersion = uint8(b)
	} else {
		return errors.Wrapf(err, "Chunk(%s): failed to read chunk version with cause", cr.target)
	}

	var chunkTimestamp time.Time
	if i64, err := utils.ReadInt64(gzRdr); err == nil {
		secs := i64 / 1000
		nanos := (i64 % 1000) * 1000000
		chunkTimestamp = time.Unix(secs, nanos)
	} else {
		return errors.Wrapf(err, "Chunk(%s): failed to read chunk timestamp with cause", cr.target)
	}
	cr.logger.Printf("Chunk(%s): version %d at time %s", cr.target, chunkVersion, chunkTimestamp)

	count := 1
	for {
		var fieldCount int32
		fieldCount, err = utils.ReadInt32(gzRdr)
		if err != nil {
			return errors.Wrapf(err,
				"Chunk(%s): failed to read field count for record %d with cause",
				cr.target, count)
		}

		rawRecord := map[string]string{}
		for ndx := int32(0); ndx < fieldCount && errors.Cause(err) != io.EOF; ndx++ {
			// we ignore each Record's 1 byte of index bit flags
			_, err = utils.ReadByte(gzRdr)
			if err != nil {
				return errors.Wrapf(err,
					"Chunk(%s): failed to read field flags for record %d with cause",
					cr.target, count)
			}

			// a Record's *key* conforms to standard Java "readUTF" behavior
			// including a max size field of 2 bytes (uint16)
			key, err := utils.ReadString(gzRdr)
			if err != nil {
				return errors.Wrapf(err,
					"Chunk(%s): failed to read field key for record %d with cause",
					cr.target, count)
			}

			// a Record's *value* can be larger; the size field is 4 bytes (int32)
			// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/Chunk.java#L189
			// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/Chunk.java#L196
			value, err := utils.ReadLargeString(gzRdr)
			if err != nil && err != io.EOF {
				return errors.Wrapf(err,
					"Chunk(%s): failed to read field value for key %s on record %d with cause",
					cr.target, key, count)
			}

			rawRecord[key] = value
		}

		record, rErr := data.NewRecord(cr.logger, rawRecord)
		if cr.filterFn != nil && !cr.filterFn(record) {
			if cr.cfg.Verbose {
				cr.logger.Printf("Chunk(%s): skipping filtered record: %+v", cr.target, record)
			}
			if errors.Cause(err) == io.EOF {
				return nil
			}
			continue
		}
		if rErr != nil {
			return errors.Wrapf(err,
				"Chunk(%s): failed to compose well-formed record %d from %s with cause: %s",
				cr.target, count, rawRecord, rErr)
		}

		cr.buffer <- record

		if errors.Cause(err) == io.EOF {
			cr.logger.Printf("Chunk: successfully published %d records from %s", count, resource)
			return nil
		}
		count++
	}
}
