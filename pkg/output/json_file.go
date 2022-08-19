package output

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

type JSON struct {
	logger   *log.Logger
	filePath string
	input    <-chan data.Record
}

func NewJSON(l *log.Logger, in <-chan data.Record, fp string) JSON {
	return JSON{l, fp, in}
}

func (j JSON) Write() error {
	var w *bufio.Writer
	if len(j.filePath) > 0 {
		path := filepath.Dir(j.filePath)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to create output directory at %s with cause", path)
		}

		f, err := os.Create(j.filePath)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to create output file at %s with cause", j.filePath)
		}
		defer f.Close()

		w = bufio.NewWriter(f)
	} else {
		w = bufio.NewWriter(os.Stdout)
	}
	defer w.Flush()

	_, err := w.WriteString("[\n")
	if err != nil {
		return errors.Wrapf(err, "JSON: failed initial write to output file %s with cause", j.filePath)
	}

	count := 0
	for record := range j.input {
		if count > 0 {
			w.WriteString(",\n")
		}

		// inject record's RecordType into payload prior to JSON  serialization
		record.Payload()["recordType"] = data.RecordTypeNames[record.Type()]

		out, err := json.Marshal(record.Payload())
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to encode Record %d to output file %s with cause", count+1, j.filePath)
		}
		_, err = w.Write(out)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to write Record %d to output file %s with cause", count+1, j.filePath)
		}
		count++
	}
	w.WriteString("\n]")

	j.logger.Printf("JSON: successfully persisted %d records to file %s", count, j.filePath)
	return nil
}
