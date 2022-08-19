package output

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

type JSON struct {
	logger *log.Logger
	cfg    config.Index
	input  <-chan data.Record
}

func NewJSON(l *log.Logger, in <-chan data.Record, c config.Index) JSON {
	l.Printf("Output: formatting data.Records as JSON...\n")
	return JSON{l, c, in}
}

func (j JSON) Write() error {
	var w *bufio.Writer
	if len(j.cfg.Output.File) > 0 {
		path := filepath.Dir(j.cfg.Output.File)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to create output directory at %s with cause", path)
		}

		f, err := os.Create(j.cfg.Output.File)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to create output file at %s with cause", j.cfg.Output.File)
		}
		defer f.Close()

		w = bufio.NewWriter(f)
	} else {
		w = bufio.NewWriter(os.Stdout)
	}
	defer w.Flush()

	_, err := w.WriteString("[\n")
	if err != nil {
		return errors.Wrapf(err, "JSON: failed initial write to output file %s with cause", j.cfg.Output.File)
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
			return errors.Wrapf(err, "JSON: failed to encode Record %d to output file %s with cause", count+1, j.cfg.Output.File)
		}
		_, err = w.Write(out)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to write Record %d to output file %s with cause", count+1, j.cfg.Output.File)
		}
		count++
	}
	w.WriteString("\n]")

	j.logger.Printf("JSON: successfully persisted %d records to file %s", count, j.cfg.Output.File)
	return nil
}
