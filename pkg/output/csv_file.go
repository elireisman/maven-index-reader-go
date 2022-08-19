package output

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
)

var (
	commaPattern    = regexp.MustCompile(`,`)
	dblQuotePattern = regexp.MustCompile(`"`)
)

type CSV struct {
	logger   *log.Logger
	filePath string
	input    <-chan data.Record
}

func NewCSV(l *log.Logger, in <-chan data.Record, fp string) CSV {
	return CSV{l, fp, in}
}

func (c CSV) Write() error {
	// TODO(eli): yuck! all this is terrible revisit the pattern, pass the writer in!
	var w *bufio.Writer
	if len(c.filePath) > 0 {
		path := filepath.Dir(c.filePath)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "JSON: failed to create output directory at %s with cause", path)
		}

		f, err := os.Create(c.filePath)
		if err != nil {
			return errors.Wrapf(err, "CSV: failed to create output file at %s with cause", c.filePath)
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
				return errors.Wrapf(err, "CSV: failed to write headers to file %s with cause", c.filePath)
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

	c.logger.Printf("CSV: successfully persisted %d records to file %s", count, c.filePath)
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
