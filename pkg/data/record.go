package data

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"

	"github.com/elireisman/maven-index-reader-go/pkg/data/types/record/keys"
)

// RecordType - the type of index entry this record represents.
// For index reading purposes, only ArtifactAdd and ArtifactRemove
// are important to parse and retain.
type RecordType uint8

// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/Record.java#L291-L365
const (
	Descriptor RecordType = iota
	ArtifactAdd
	ArtifactRemove
	AllGroups
	RootGroups
)

// ArtifactAddRecordKeys - determines an order on an ARTIFACT_ADD
// Record's keys that callers can rely on for output formats
// that require it.
var ArtifactAddRecordKeys = []keys.Record{
	keys.RecordModified,
	keys.GroupID,
	keys.ArtifactID,
	keys.Version,
	keys.Classifier,
	keys.FileExtension,
	keys.FileModified,
	keys.FileSize,
	keys.Packaging,
	keys.HasSources,
	keys.HasJavadoc,
	keys.HasSignature,
	keys.Name,
	keys.Description,
	keys.SHA1,
	keys.Classnames,
	keys.PluginPrefix,
	keys.PluginGoals,
}

// ArtifactRemoveRecordKeys - determines an order on an ARTIFACT_REMOVED
// Record's keys that callers can rely on for output formats that
// require it.
var ArtifactRemoveRecordKeys = []keys.Record{
	keys.RecordModified,
	keys.GroupID,
	keys.ArtifactID,
	keys.Version,
	keys.Classifier,
	keys.FileExtension,
	keys.Packaging,
}

// Record -
type Record struct {
	kind RecordType
	data map[keys.Record]interface{}
	keys []keys.Record
}

// Type - expose index record types for callers
func (r Record) Type() RecordType {
	return r.kind
}

// Keys - obtain a fixed-order list of well-formed Record attribute keys
// the caller can iterate on to obtain similarly-ordered values.
func (r Record) Keys() []keys.Record {
	return r.keys
}

// Get - obtain a parsed well-formed attribute value from the Record, or an error
func (r Record) Get(key keys.Record) (fmt.Stringer, error) {
	val, found := r.data[key]
	if !found {
		return nil, errors.Errorf("Record#Get: no value found for key %s", key)
	}

	// TODO(eli): this is probably a terrible idea, FIX IT!
	s, ok := val.(fmt.Stringer)
	if !ok {
		return nil, errors.Errorf("Record#Get: value does not implement fmt.Stringer for key %s", key)
	}

	return s, nil
}

// NewRecord - parses the input map then populates
// and returns a well-formed Record, or an error.
func NewRecord(logger *log.Logger, raw map[string]string) (Record, error) {
	// at the moment, no effort is made to parse any of these
	// Record types any farther.
	if _, found := raw[keys.Descriptor]; found {
		return Record{kind: Descriptor}, nil
	}
	if _, found := raw[keys.AllGroups]; found {
		return Record{kind: AllGroups}, nil
	}
	if _, found := raw[keys.RootGroups]; found {
		return Record{kind: RootGroups}, nil
	}

	// attempt to parse a well-formed ARTFACT_REMOVE
	if _, found := raw[keys.Del]; found {
		return newArtifactRecord(raw, ArtifactRemove, ArtifactRemoveRecordKeys)
	}

	// attempt to parse a well-formed ARTIFACT_ADD
	return newArtifactRecord(raw, ArtifactAdd, ArtifactAddRecordKeys)
}

func newArtifactRecord(raw map[string]string, artifactType RecordType, artifactKeys []keys.Record) (Record, error) {
	out := Record{
		kind: artifactType,
		data: map[keys.Record]interface{}{},
		keys: artifactKeys,
	}

	// hacky fix on raw input data for known write-time bug on index source
	patchForMIndexer41Bug(artifactType, raw)

	// split out UINFO values once and pass into the per-key parser to save work
	uinfoVals := splitValue(raw[UInfoKey])

	// check the raw input record for each registered key and value,
	// storing a parsed value or default on the output Record
	for i := 0; i < len(out.keys); i++ {
		key := out.keys[i]
		rawValue, found := raw[string(key)]
		if found {
			value, err := parseRecordValue(key, rawValue, uinfoVals)
			if err != nil {
				return out, errors.Wrapf(err,
					"Record: failed to parse key %s from %s with cause",
					key, rawValue)
			}
			out.data[key] = value
		}
	}

	return out, nil
}

const (
	// Raw key on "ARTIFACT_ADD" type Records
	UInfoKey = "u"

	// Raw key on "ARTIFACT_ADD" type Records
	InfoKey = "i"

	// Parsed value for null records found in index source
	NotAvailable = "NA"

	// Field separator for parsing raw index-sourced records with multi values
	RecordValueSeparator = `|`
)

// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/RecordExpander.java#L66-L81
func patchForMIndexer41Bug(artifactType RecordType, raw map[string]string) {
	if artifactType != ArtifactAdd {
		return
	}

	rawUInfo, uFound := raw[UInfoKey]
	rawInfo, iFound := raw[InfoKey]

	if uFound && iFound && len(strings.Trim(rawInfo, " \t\n\r")) > 0 {
		vals := splitValue(rawInfo)
		if len(vals) > 6 {
			if strings.HasSuffix(rawUInfo, RecordValueSeparator+NotAvailable) {
				fileExt := vals[6]
				raw[UInfoKey] = rawUInfo + RecordValueSeparator + fileExt
			}
		}
	}
}

// TODO(eli): implement this; check return type!
func expandUInfo(rawValue map[string]string) []string {
	// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/RecordExpander.java#L188-L212
	return nil
}

// TODO(eli): Process raw string value from index into well-formed Record value
func parseRecordValue(key keys.Record, rawValue string, uinfoVals []string) (interface{}, error) {
	var out interface{}

	// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/RecordExpander.java#L118-L185

	switch key {
	case keys.RecordModified:
	case keys.GroupID:
	case keys.ArtifactID:
	case keys.Version:
	case keys.Classifier:
	case keys.FileExtension:
	case keys.FileModified:
	case keys.FileSize:
	case keys.Packaging:
	case keys.HasSources:
	case keys.HasJavadoc:
	case keys.HasSignature:
	case keys.Name:
	case keys.Description:
	case keys.SHA1:
	case keys.Classnames:
	case keys.PluginPrefix:
	case keys.PluginGoals:
	default:
	}

	return out, nil
}

func splitValue(rawValue string) []string {
	var out []string
	for _, elem := range strings.Split(rawValue, RecordValueSeparator) {
		if len(elem) > 0 {
			out = append(out, elem)
		}
	}

	return out
}
