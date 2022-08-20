package output

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

type CSV struct {
	logger *log.Logger
	cfg    config.Index
	input  <-chan data.Record
}

func NewCSV(l *log.Logger, in <-chan data.Record, c config.Index) CSV {
	l.Printf("Output: formatting data.Records as CSV...\n")
	return CSV{l, c, in}
}

func (c CSV) Write() error {
	var w *csv.Writer
	if len(c.cfg.Output.File) > 0 {
		path := filepath.Dir(c.cfg.Output.File)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "CSV: failed to create output directory at %s with cause", path)
		}

		f, err := os.Create(c.cfg.Output.File)
		if err != nil {
			return errors.Wrapf(err, "CSV: failed to create output file at %s with cause", c.cfg.Output.File)
		}
		defer f.Close()

		w = csv.NewWriter(f)
	} else {
		w = csv.NewWriter(os.Stdout)
	}
	defer w.Flush()

	count := 0
	headersWritten := false
	for record := range c.input {
		if !headersWritten {
			// obtain ordered list of keys, prefixed with RecordType
			keys := []string{"record_type"}
			keys = append(keys, record.Keys()...)

			if err := w.Write(keys); err != nil {
				return errors.Wrapf(err, "CSV: failed to write headers to file %s with cause", c.cfg.Output.File)
			}
			headersWritten = true
		}

		// append data.Record's RecordType as 1st value
		var values []string
		values = append(values, data.RecordTypeNames[record.Type()])

		// collect remaining Record values in key-order
		for i := 0; i < len(record.Keys()); i++ {
			key := record.Keys()[i]
			value := ""

			rawValue := record.Get(key)
			if rawValue != nil {
				value = fmt.Sprintf("%v", rawValue)
			}
			values = append(values, value)
		}

		if err := w.Write(values); err != nil {
			return errors.Wrap(err, "CSV: failed to write values to file %s with cause")
		}

		count++
	}

	c.logger.Printf("CSV: successfully persisted %d records to file %s", count, c.cfg.Output.File)
	return nil
}
