package config

import (
	"fmt"
	"log"
)

// Validate - the beginnings of a config.Index validator :)
func Validate(logger *log.Logger, cfg Index) error {
	logger.Printf("Resolved configuration: %+v\n", cfg)

	if len(cfg.Meta.File) == 0 {
		return fmt.Errorf("Invalid configuration: index base file name (Meta.File) is required")
	}
	if len(cfg.Meta.ID) == 0 {
		return fmt.Errorf("Invalid configuration: index identifier (Meta.ID) is required")
	}
	if len(cfg.Meta.ChainID) == 0 {
		return fmt.Errorf("Invalid configuration: index chain ID (Meta.ChainID) is required")
	}
	if len(cfg.Source.Base) == 0 {
		return fmt.Errorf("Invalid configuration: index base URL (Source.Base) is required")
	}
	if cfg.Source.Type != Local && cfg.Source.Type != HTTP {
		return fmt.Errorf("Invalid configuration: index location (Source.Type) is required")
	}
	if cfg.Output.Format != Log && cfg.Output.Format != CSV && cfg.Output.Format != JSON {
		return fmt.Errorf("Invalid configuration: valid format type (Output.Format) is required")
	}

	return nil
}

// configuration for an readers.IndexReader
type Index struct {
	Meta   Meta
	Source Source
	Mode   Mode
	Output Output
}

// Resolve the full Resource target string from supplied config.Index and args
func (cfg Index) ResolveTarget(targetOrPattern string, targetArgs ...interface{}) string {
	return cfg.Source.Base + cfg.Meta.File + fmt.Sprintf(targetOrPattern, targetArgs...)
}

type Meta struct {
	ID      string // expected index ID, as in "nexus.index.id"
	ChainID string // expected chain ID, as in "nexus.index.chain-id"
	File    string // expected base name of source index resources like "nexus-maven-repository-index"
}

type Mode struct {
	// one of 'all', 'from-time' or 'from-chunk'
	Type ModeType
	// Unix millis since or chunk ID of last successfully
	// ingested incremental chunk, depending on ModeType
	From int64
}

func (m Mode) Incremental() bool {
	return m.Type != All && m.From > 0
}

type ModeType uint8

const (
	UnknownMode ModeType = iota
	All
	FromTime
	FromChunk
)

var ModeTypes = map[string]ModeType{
	"all":        All,
	"from-time":  FromTime,
	"from-chunk": FromChunk,
}

type Source struct {
	Base string     // either the base URL or absolute base path depending on SourceType
	Type SourceType // enum of local filesystem or HTTP based index source types
}

type SourceType uint8

const (
	UnknownSource SourceType = iota
	Local
	HTTP
)

type Output struct {
	Format OutputType
	File   string // defaults to os.Stdout if undefined
}

type OutputType uint8

const (
	UnknownOuput OutputType = iota
	Log
	CSV
	JSON
)

var OutputFormats = map[string]OutputType{
	"log":  Log,
	"csv":  CSV,
	"json": JSON,
}
