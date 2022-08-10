package readers

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elireisman/maven-index-reader-go/internal/util"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

var linesPattern = regexp.MustCompile(`[\r\n]+`)

type PropertiesReader struct {
	logger     *log.Logger
	resource   resources.Resource
	properties map[string]string
}

func NewProperties(l *log.Logger, r resources.Resource) (*PropertiesReader, error) {
	return &PropertiesReader{
		logger:     l,
		resource:   r,
		properties: map[string]string{},
	}, nil
}

func (pr *PropertiesReader) Execute() error {
	if err := pr.loadProperties(); err != nil {
		return errors.Wrap(err, "PropertiesReader: failed to parse Java properties from data with cause")
	}

	for k, v := range pr.properties {
		pr.logger.Printf("%24s => %40s\n", k, v)
	}

	return nil
}

func (pr *PropertiesReader) GetAsString(key string) (string, error) {
	val, found := pr.properties[key]
	if !found {
		return "", errors.Errorf("GetAsString: expected properties key %q not found", key)
	}

	return val, nil
}

func (pr *PropertiesReader) GetAsInt(key string) (int, error) {
	val, found := pr.properties[key]
	if !found {
		return 0, errors.Errorf("GetAsInt: expected properties key %q not found", key)
	}

	out, err := strconv.Atoi(val)
	if err != nil {
		return 0, errors.Wrapf(err, "GetAsInt: failed to parse expected integer value %q for key %q with cause", key, val)
	}

	return out, nil
}

func (pr *PropertiesReader) GetAsTimestamp(key string) (time.Time, error) {
	val, found := pr.properties[key]
	if !found {
		return time.Now().UTC(), errors.Errorf("GetAsTimestamp: expected properties key %q not found", key)
	}

	t, err := util.GetTimestamp(val)
	if err != nil {
		return time.Now().UTC(), errors.Wrapf(err, "GetAsTimestamp: failed to parse expected time value %q for key %q with cause", key, val)
	}

	return t, nil
}

// Extract Java-style properties file from io.Reader and build
// a map of "raw" string keys and values for the caller to
// introspect on. Typically used to identify incremental index
// files of interest when performing a full backfill or incremental
// update.
func (pr *PropertiesReader) loadProperties() error {
	rdr, err := pr.resource.Reader()
	if err != nil {
		return errors.Wrap(err, "PropertiesReader: failed to read data from Resource with cause")
	}

	raw, err := ioutil.ReadAll(rdr)
	if err != nil {
		return errors.Wrap(err, "PropertiesReader: failed to read raw data from input with cause")
	}

	pr.properties = map[string]string{}
	for ndx, line := range linesPattern.Split(string(raw), -1) {
		// skip commented out lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return errors.Errorf("PropertiesReader: line %d failed to parse into key and value: %s", ndx, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		pr.properties[key] = value
	}

	return nil
}
