package config

import "fmt"

// configuration for an readers.IndexReader
type Index struct {
	Meta   Meta
	Source Source
	Mode   Mode
}

// Resolve the full Resource target string from supplied config.Index and args
func (cfg Index) ResolveTarget(targetOrPattern string, targetArgs ...interface{}) string {
	return cfg.Source.Base + cfg.Meta.Target + fmt.Sprintf(targetOrPattern, targetArgs...)
}

type Meta struct {
	ID      string // expected index ID, as in "nexus.index.id"
	ChainID string // expected chain ID, as in "nexus.index.chain-id"
	Target  string // expected base name of source index resources like "nexus-maven-repository-index"
}

type Mode struct {
	Incremental bool
	FromChunk   int // fetch all available chunks with this ID and higher
	//FromTime    time.Time // fetch all available chunks with this timestamp or more recent
}

type Source struct {
	Base string     // either the base URL or absolute base path depending on SourceType
	Type SourceType // enum of local filesystem or HTTP based index source types
}

type SourceType uint8

const (
	Unknown SourceType = iota
	Local
	HTTP
)
