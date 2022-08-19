package data

import (
	"log"
	"strconv"
	"strings"

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

var RecordTypeNames = map[RecordType]string{
	Descriptor:     "descriptor",
	ArtifactAdd:    "artifact_add",
	ArtifactRemove: "artifact_remove",
	AllGroups:      "all_groups",
	RootGroups:     "root_groups",
}

const (
	// Raw key on "ARTIFACT_ADD" type Records
	UInfoKey = "u"

	// Raw key on "ARTIFACT_ADD" type Records
	InfoKey = "i"

	// Raw key on "ARTIFACT_ADD" type Records
	NameKey = "n"

	// Raw key on "ARTIFACT_ADD" type Records
	DescriptionKey = "d"

	// Raw key on "ARTIFACT_ADD" type Records
	RecordModifiedKey = "m"

	// Raw key on "ARTIFACT_ADD" type Records
	ClassnamesKey = "classnames"

	// Raw key on "ARTIFACT_ADD" type Records
	SHA1Key = "1"

	// Parsed value for null records found in index source
	NotAvailable = "NA"

	// Field separator for parsing raw index-sourced records with multi values
	RecordValueSeparator = `|`
)

// The RecordKey lists below determine an ordering for iterating over
// known Record values, for Output formats that require stable ordering
var (
	ArtifactAddRecordKeys = []keys.Record{
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

	ArtifactRemoveRecordKeys = []keys.Record{
		keys.RecordModified,
		keys.GroupID,
		keys.ArtifactID,
		keys.Version,
		keys.Classifier,
		keys.FileExtension,
		keys.Packaging,
	}

	DescriptorRecordKeys = []keys.Record{
		keys.RepositoryID,
	}

	AllGroupsRecordKeys = []keys.Record{
		keys.AllGroupsList,
	}

	RootGroupsRecordKeys = []keys.Record{
		keys.RootGroupsList,
	}
)

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

// Get - obtain a parsed well-formed, well-typed attribute value
// as stored on the given Record, or nil if no value is present.
// result values must be cast by caller if non-nil
func (r Record) Get(key keys.Record) interface{} {
	val, found := r.data[key]
	if !found {
		return nil
	}

	return val
}

// Payload - obtain the full internal representation of the Record's
// data attributes. Useful for output formats that can more easily work
// with this data
func (r Record) Payload() map[keys.Record]interface{} {
	return r.data
}

// NewRecord - parses the input map then populates
// and returns a well-formed Record, or an error.
func NewRecord(logger *log.Logger, indexRecord map[string]string) (Record, error) {
	if _, found := indexRecord[keys.Descriptor]; found {
		return newDescriptorRecord(indexRecord)
	}
	if _, found := indexRecord[keys.AllGroups]; found {
		return newAllGroupsRecord(indexRecord)
	}
	if _, found := indexRecord[keys.RootGroups]; found {
		return newRootGroupsRecord(indexRecord)
	}
	if _, found := indexRecord[keys.Del]; found {
		return newArtifactRemoveRecord(indexRecord)
	}

	return newArtifactAddRecord(indexRecord)
}

func newDescriptorRecord(indexRecord map[string]string) (Record, error) {
	out := Record{
		kind: Descriptor,
		data: map[keys.Record]interface{}{},
		keys: DescriptorRecordKeys,
	}

	if rawIDXInfoVal, found := indexRecord["IDXINFO"]; found {
		splits := splitValue(rawIDXInfoVal)
		out.data[keys.RepositoryID] = splits[1]
	}

	return out, nil
}

func newAllGroupsRecord(indexRecord map[string]string) (Record, error) {
	out := Record{
		kind: AllGroups,
		data: map[keys.Record]interface{}{},
		keys: AllGroupsRecordKeys,
	}

	if groups, ok := stringArrayIfNotNull(indexRecord, string(keys.AllGroupsList)); ok {
		out.data[keys.AllGroupsList] = groups
	}

	return out, nil
}

func newRootGroupsRecord(indexRecord map[string]string) (Record, error) {
	out := Record{
		kind: RootGroups,
		data: map[keys.Record]interface{}{},
		keys: RootGroupsRecordKeys,
	}

	if groups, ok := stringArrayIfNotNull(indexRecord, string(keys.RootGroupsList)); ok {
		out.data[keys.RootGroupsList] = groups
	}

	return out, nil
}

func newArtifactRemoveRecord(indexRecord map[string]string) (Record, error) {
	out := Record{
		kind: ArtifactRemove,
		data: map[keys.Record]interface{}{},
		keys: ArtifactRemoveRecordKeys,
	}

	// populate fields using index source internal fields
	if tsMillis, ok := millisTimestampIfNotNull(indexRecord, "m"); ok {
		out.data[keys.RecordModified] = tsMillis
	}
	out = expandUInfo(indexRecord, string(keys.Del), out)

	return out, nil
}

func newArtifactAddRecord(indexRecord map[string]string) (Record, error) {
	out := Record{
		kind: ArtifactAdd,
		data: map[keys.Record]interface{}{},
		keys: ArtifactAddRecordKeys,
	}

	// hacky fix on raw input data for known write-time bug on index source
	patchForMIndexer41Bug(indexRecord)

	// populate basic fields on data.Record from index source
	out = expandUInfo(indexRecord, UInfoKey, out)
	rawInfo, found := indexRecord[InfoKey]
	if found && len(rawInfo) > 0 {
		vals := splitValue(rawInfo)

		// packaging type if present ("jar", "war", "ear", etc.)
		if vals[0] == NotAvailable {
			out.data[keys.Packaging] = nil
		}
		out.data[keys.Packaging] = vals[0]

		// int64 as Java "Unix" milliseconds timestamp
		fm, err := strconv.ParseInt(vals[1], 10, 64)
		if err != nil {
			fm = 0
		}
		out.data[keys.FileModified] = fm

		// int64 as file size
		fs, err := strconv.ParseInt(vals[2], 10, 64)
		if err != nil {
			fs = 0
		}
		out.data[keys.FileSize] = fs

		out.data[keys.HasSources] = strings.Trim(vals[3], " \t\r\n") == "1"
		out.data[keys.HasJavadoc] = strings.Trim(vals[4], " \t\r\n") == "1"
		out.data[keys.HasSignature] = strings.Trim(vals[5], " \t\r\n") == "1"
		if len(vals) > 6 {
			out.data[keys.FileExtension] = vals[6]
		} else {
			pkgVal := vals[0]
			if vals[0] != NotAvailable {
				_, foundClassifier := out.data[keys.Classifier]
				if foundClassifier || pkgVal == "pom" || pkgVal == "war" || pkgVal == "ear" {
					out.data[keys.FileExtension] = pkgVal
				} else {
					// default value; best guess
					out.data[keys.FileExtension] = "jar"
				}
			}
		}
	}

	// populate other fields using index source internal fields
	if tsMillis, ok := millisTimestampIfNotNull(indexRecord, RecordModifiedKey); ok {
		out.data[keys.RecordModified] = tsMillis
	}
	if name, ok := stringIfNotNull(indexRecord, NameKey); ok {
		out.data[keys.Name] = name
	}
	if desc, ok := stringIfNotNull(indexRecord, DescriptionKey); ok {
		out.data[keys.Description] = desc
	}
	if sha1, ok := stringIfNotNull(indexRecord, SHA1Key); ok {
		out.data[keys.Description] = sha1
	}
	if classNames, ok := stringArrayIfNotNull(indexRecord, ClassnamesKey); ok {
		out.data[keys.Classnames] = classNames
	}

	// populate OPTIONAL Maven Plugin fields, if present
	if plgPrf, ok := stringIfNotNull(indexRecord, "px"); ok {
		out.data[keys.PluginPrefix] = plgPrf
	}
	if plgGoals, ok := stringIfNotNull(indexRecord, "gx"); ok {
		out.data[keys.PluginGoals] = plgGoals
	}

	// populate OPTIONAL OSGI fields, if present
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleSymbolicName)); ok {
		out.data[keys.OSGIBundleSymbolicName] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleVersion)); ok {
		out.data[keys.OSGIBundleVersion] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIExportPackage)); ok {
		out.data[keys.OSGIExportPackage] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIExportService)); ok {
		out.data[keys.OSGIExportService] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleDescription)); ok {
		out.data[keys.OSGIBundleDescription] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleName)); ok {
		out.data[keys.OSGIBundleName] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleLicense)); ok {
		out.data[keys.OSGIBundleLicense] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBundleDocURL)); ok {
		out.data[keys.OSGIBundleDocURL] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIImportPackage)); ok {
		out.data[keys.OSGIImportPackage] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIRequireBundle)); ok {
		out.data[keys.OSGIRequireBundle] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIProvideCapability)); ok {
		out.data[keys.OSGIProvideCapability] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIRequireCapability)); ok {
		out.data[keys.OSGIRequireCapability] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIFragmentHost)); ok {
		out.data[keys.OSGIFragmentHost] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGIBREE)); ok {
		out.data[keys.OSGIBREE] = v
	}
	if v, ok := stringIfNotNull(indexRecord, string(keys.OSGISHA256)); ok {
		out.data[keys.OSGISHA256] = v
	}

	return out, nil
}

func stringIfNotNull(indexRecord map[string]string, rawKey string) (string, bool) {
	rawValue, found := indexRecord[rawKey]
	return rawValue, (found && len(strings.Trim(rawValue, " \t\r\n")) > 0)
}

func stringArrayIfNotNull(indexRecord map[string]string, rawKey string) ([]string, bool) {
	rawValue, found := indexRecord[rawKey]
	if found && len(strings.Trim(rawValue, " \t\r\n")) > 0 {
		return splitValue(rawValue), true
	}

	return nil, false
}

func millisTimestampIfNotNull(indexRecord map[string]string, rawKey string) (int64, bool) {
	rawValue, found := indexRecord[rawKey]
	if found && len(rawValue) > 0 {
		if tsMillis, err := strconv.ParseInt(rawValue, 10, 64); err != nil {
			return tsMillis, true
		}
	}

	return 0, false
}

// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/RecordExpander.java#L66-L81
func patchForMIndexer41Bug(indexRecord map[string]string) {
	rawUInfo, uFound := indexRecord[UInfoKey]
	rawInfo, iFound := indexRecord[InfoKey]

	if uFound && iFound && len(strings.Trim(rawInfo, " \t\n\r")) > 0 {
		vals := splitValue(rawInfo)
		if len(vals) > 6 {
			if strings.HasSuffix(rawUInfo, RecordValueSeparator+NotAvailable) {
				fileExt := vals[6]
				indexRecord[UInfoKey] = rawUInfo + RecordValueSeparator + fileExt
			}
		}
	}
}

// https://github.com/apache/maven-indexer/blob/31052fdeebc8a9f845eb18cd4c13669b316b3e29/indexer-reader/src/main/java/org/apache/maven/index/reader/RecordExpander.java#L188-L212
func expandUInfo(indexRecord map[string]string, uinfoKey string, out Record) Record {
	rawUInfo, uFound := indexRecord[uinfoKey]
	if uFound && len(rawUInfo) > 0 {
		vals := splitValue(rawUInfo)
		out.data[keys.GroupID] = vals[0]
		out.data[keys.ArtifactID] = vals[1]
		out.data[keys.Version] = vals[2]

		if len(vals) > 3 && len(vals[3]) > 0 && vals[3] != NotAvailable {
			out.data[keys.Classifier] = vals[3]
			if len(vals) > 4 {
				out.data[keys.FileExtension] = vals[4]
			}
		} else if len(vals) > 4 {
			out.data[keys.Packaging] = vals[4]
		}
	}

	return out
}

func splitValue(rawValue string) []string {
	var out []string

	splits := strings.Split(rawValue, RecordValueSeparator)
	for i := 0; i < len(splits); i++ {
		elem := splits[i]
		out = append(out, elem)
	}

	return out
}
