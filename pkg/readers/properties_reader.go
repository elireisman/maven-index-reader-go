package readers

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

var linesPattern = regexp.MustCompile(`[\r\n]+`)

type PropertiesReader struct {
	logger   *log.Logger
	resource resources.Resource
}

func NewProperties(l *log.Logger, r resources.Resource) (PropertiesReader, error) {
	return PropertiesReader{
		logger:   l,
		resource: r,
	}, nil
}

// Extract Java-style properties file from io.Reader and build
// a map of "raw" string keys and values for the caller to
// introspect on. Typically used to identify incremental index
// files of interest when performing a full backfill or incremental
// update.
func (pr PropertiesReader) Read() (data.Properties, error) {
	pr.logger.Printf("PropertiesReader: consuming %+v", pr.resource)
	rdr, err := pr.resource.Reader()
	if err != nil {
		return data.Properties{}, errors.Wrap(err, "PropertiesReader: failed to read data from Resource with cause")
	}
	defer pr.resource.Close()

	raw, err := ioutil.ReadAll(rdr)
	if err != nil {
		return data.Properties{}, errors.Wrap(err, "PropertiesReader: failed to read raw data from input with cause")
	}

	goStr, err := utils.GetString(raw)
	if err != nil {
		return data.Properties{}, errors.Wrap(err, "PropertiesReader: failed to convert data bytes from Java 'modified' UTF-8 with cause")
	}

	out := map[string]string{}
	for ndx, line := range linesPattern.Split(goStr, -1) {
		// skip commented out lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return data.Properties{}, errors.Errorf("PropertiesReader: line %d failed to parse into key and value: %s", ndx, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		pr.logger.Printf("%16s => %24s", key, value)
		out[key] = value
	}

	return data.NewProperties(out), nil
}
