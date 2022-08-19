package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

var (
	commaPattern    = regexp.MustCompile(`,`)
	dblQuotePattern = regexp.MustCompile(`"`)
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
	var w *bufio.Writer
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

		w = bufio.NewWriter(f)
	} else {
		w = bufio.NewWriter(os.Stdout)
	}
	defer w.Flush()

	count := 0
	headersWritten := false
	for record := range c.input {
		var values []string

		if !headersWritten {
			// ordered array or strings
			keys := record.Keys()
			_, err := w.WriteString("record_type," + strings.Join(keys, ",") + "\n")
			if err != nil {
				return errors.Wrapf(err, "CSV: failed to write headers to file %s with cause", c.cfg.Output.File)
			}
			headersWritten = true
		}

		// append data.Record's RecordType as 1st value
		values = append(values, data.RecordTypeNames[record.Type()])

		for i := 0; i < len(record.Keys()); i++ {
			key := record.Keys()[i]
			value := record.Get(key)

			// TODO(eli): each "raw" value will need to be:
			// 1. Scanned and escaped for CSV separator values!
			// 2. Multiple-entry values ([]string, etc.) must be formatted properly etc.
			formattedValue := ""
			if value != nil {
				formattedValue = escapeForCSV(value)
			}
			values = append(values, formattedValue)
		}

		_, err := w.WriteString(strings.Join(values, ",") + "\n")
		if err != nil {
			return errors.Wrap(err, "CSV: failed to write values to file %s with cause")
		}

		count++
	}

	c.logger.Printf("CSV: successfully persisted %d records to file %s", count, c.cfg.Output.File)
	return nil
}

func escapeForCSV(value interface{}) string {
	raw := fmt.Sprintf("%v", value)

	if strings.ContainsAny(raw, `",`) {
		escaped := dblQuotePattern.ReplaceAllString(raw, `"""`)
		return `"` + commaPattern.ReplaceAllString(escaped, `","`) + `"`
	}

	return raw
}
