package output

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

type CSV struct {
	logger   *log.Logger
	filePath string
	input    <-chan data.Record
}

func NewCSV(l *log.Logger, fp string, in <-chan data.Record) CSV {
	return CSV{l, fp, in}
}

func (c CSV) Write() error {
	f, err := os.Create(c.filePath)
	if err != nil {
		return errors.Wrapf(err, "CSV: failed to create output file at %s with cause", c.filePath)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	count := 0
	headersWritten := false
	for record := range c.input {
		var values []string

		if !headersWritten {
			// ordered array or strings
			keys := record.Keys()
			_, err := w.WriteString(strings.Join(keys, ",") + "\n")
			if err != nil {
				return errors.Wrapf(err, "CSV: failed to write headers to file %s with cause", c.filePath)
			}
			headersWritten = true
		}

		for i := 0; i < len(record.Keys()); i++ {
			key := record.Keys()[i]
			value, err := record.Get(key)
			if err != nil {
				return errors.Wrapf(err, "CSV: failed to resolve expected Record value for key %s with cause", key)
			}

			// TODO(eli): each "raw" value will need to be:
			// 1. Scanned and escaped for CSV separator values!
			// 2. Multiple-values must be formatted properly etc.
			formattedValue := ""
			if value != nil {
				formattedValue = value.String()
			}
			values = append(values, formattedValue)
		}

		_, err := w.WriteString(strings.Join(values, ",") + "\n")
		if err != nil {
			return errors.Wrap(err, "CSV: failed to write values to file %s with cause")
		}

		count++
	}

	c.logger.Printf("CSV: successfully persisted %d records to file %s", count, c.filePath)
	return nil
}
