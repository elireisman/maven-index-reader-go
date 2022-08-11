package data

import (
	"strconv"
	"time"

	"github.com/elireisman/maven-index-reader-go/internal/utils"

	"github.com/pkg/errors"
)

// Properties represents a well-formed Java properties file
// along with some convenience methods.
type Properties struct {
	properties map[string]string
}

func NewProperties(props map[string]string) Properties {
	return Properties{props}
}

func (pr Properties) GetAsString(key string) (string, error) {
	val, found := pr.properties[key]
	if !found {
		return "", errors.Errorf("GetAsString: expected properties key %q not found", key)
	}

	return val, nil
}

func (pr Properties) GetAsInt(key string) (int, error) {
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

func (pr Properties) GetAsTimestamp(key string) (time.Time, error) {
	val, found := pr.properties[key]
	if !found {
		return time.Now().UTC(), errors.Errorf("GetAsTimestamp: expected properties key %q not found", key)
	}

	t, err := utils.GetTimestamp(val)
	if err != nil {
		return time.Now().UTC(), errors.Wrapf(err, "GetAsTimestamp: failed to parse expected time value %q for key %q with cause", key, val)
	}

	return t, nil
}
