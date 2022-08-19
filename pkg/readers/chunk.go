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
	suffix string
	cfg    config.Index
	logger *log.Logger
	buffer chan<- data.Record
}

// NewChunk - caller supplies the input resource as well as the
// output channel for captured records that the caller plans to consume
func NewChunk(l *log.Logger, b chan<- data.Record, c config.Index, s string) Chunk {
	return Chunk{
		suffix: s,
		cfg:    c,
		logger: l,
		buffer: b,
	}
}

// Read - initiate async consumption of Resource and population of data.Record buffer
func (cr Chunk) Read() error {
	resource, err := resources.ConfigureResource(cr.logger, cr.cfg, cr.suffix)
	rdr, err := resource.Reader()
	if err != nil {
		return errors.Wrapf(err, "Chunk: failed to obtain data stream from %s with cause", resource)
	}
	defer resource.Close()

	gzRdr, err := gzip.NewReader(rdr)
	if err != nil {
		return errors.Wrapf(err, "Chunk: failed to wrap %s in GZIP Reader with cause", resource)
	}
	defer gzRdr.Close()

	var chunkVersion uint8
	if b, err := utils.ReadByte(gzRdr); err == nil {
		chunkVersion = uint8(b)
	} else {
		return errors.Wrap(err, "Chunk: failed to read chunk version with cause")
	}

	var chunkTimestamp time.Time
	if i64, err := utils.ReadInt64(gzRdr); err == nil {
		secs := i64 / 1000
		nanos := (i64 % 1000) * 1000000
		chunkTimestamp = time.Unix(secs, nanos)
	} else {
		return errors.Wrap(err, "Chunk: failed to read chunk timestamp with cause")
	}
	cr.logger.Printf("Chunk: found %s of version %d at time %s", resource, chunkVersion, chunkTimestamp)

	count := 1
	for {
		var fieldCount int32
		fieldCount, err = utils.ReadInt32(gzRdr)
		if err != nil {
			return errors.Wrapf(err,
				"Chunk: failed to read field count for record %d from %s with cause",
				count, resource)
		}

		rawRecord := map[string]string{}
		for ndx := int32(0); ndx < fieldCount; ndx++ {
			// we ignore each Record's 1 byte of index bit flags
			_, err = utils.ReadByte(gzRdr)
			if err != nil {
				return errors.Wrapf(err,
					"Chunk: failed to read field flags for record %d from %s with cause",
					count, resource)
			}

			// a Record's *key* conforms to standard Java "readUTF" behavior
			// including a max size field of 2 bytes (uint16)
			key, err := utils.ReadString(gzRdr)
			if err != nil {
				return errors.Wrapf(err,
					"Chunk: failed to read field key for record %d from %s with cause",
					count, resource)
			}

			// a Record's *value* can be larger; the size field is 4 bytes (int32)
			// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/Chunk.java#L189
			// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/Chunk.java#L196
			value, err := utils.ReadLargeString(gzRdr)
			if err != nil {
				return errors.Wrapf(err,
					"Chunk: failed to read field value for key %s on record %d from %s with cause",
					key, count, resource)
			}

			rawRecord[key] = value
		}

		record, rErr := data.NewRecord(cr.logger, rawRecord)
		if isSkippableRecordType(record) {
			cr.logger.Printf("Chunk: skipped Record by type %+v", record)
			if errors.Cause(err) == io.EOF {
				return nil
			}
			continue
		}
		if rErr != nil {
			cr.logger.Panicf(
				"Chunk: failed to compose well-formed record %d from %s from %s with cause: %s (at EOF: %t)",
				count, rawRecord, resource, rErr, errors.Cause(err) == io.EOF)
		}

		cr.buffer <- record

		if errors.Cause(err) == io.EOF {
			cr.logger.Printf("Chunk: successfully published %d records from %s", count, resource)
			return nil
		}
		count++
	}
}

func isSkippableRecordType(record data.Record) bool {
	return record.Type() != data.ArtifactAdd &&
		record.Type() != data.ArtifactRemove
}
