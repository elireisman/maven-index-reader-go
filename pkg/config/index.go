package config

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Validate - the beginnings of a config.Index validator :)
func Validate(logger *log.Logger, cfg Index) error {
	logger.Printf("Resolved configuration: %+v\n", cfg)

	if len(cfg.Meta.File) == 0 {
		return errors.Errorf("Invalid configuration: index base file name (Meta.File) is required")
	}
	if len(cfg.Meta.ID) == 0 {
		return errors.Errorf("Invalid configuration: index identifier (Meta.ID) is required")
	}
	if len(cfg.Meta.ChainID) == 0 {
		return errors.Errorf("Invalid configuration: index chain ID (Meta.ChainID) is required")
	}
	if len(cfg.Source.Base) == 0 {
		return errors.Errorf("Invalid configuration: index base URL (Source.Base) is required")
	}
	if cfg.Source.Type != Local && cfg.Source.Type != HTTP {
		return errors.Errorf("Invalid configuration: index location (Source.Type) is required")
	}
	if cfg.Output.Format != Log && cfg.Output.Format != CSV && cfg.Output.Format != JSON {
		return errors.Errorf("Invalid configuration: valid format type (Output.Format) is required")
	}
	if cfg.Mode.Type > All && len(cfg.Mode.After) == 0 {
		return errors.New("Invalid configuration: Mode.Type specifies incremental run but Mode.After empty")
	}
	switch cfg.Mode.Type {
	case AfterChunk:
		if _, err := strconv.Atoi(cfg.Mode.After); err != nil {
			return errors.Wrapf(err, "Invalid configuration: invalid integer on Mode.After (%s) with cause", cfg.Mode.After)
		}

	case AfterTime:
		if _, err := time.Parse(time.RFC3339, cfg.Mode.After); err != nil {
			return errors.Wrapf(err, "Invalid configuration: invalid timestamp on Mode.After (%s) with cause", cfg.Mode.After)
		}

	case All:
		// nothing to validate here

	default:
		return errors.Errorf("Invalid configuration: invalid Mode.Type")
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
	// one of 'all', 'after-time' or 'after-chunk'
	Type ModeType
	// time.Time string since, or chunk ID of, last successfully
	// ingested incremental chunk, depending on specified ModeType
	After string
}

func (m Mode) Incremental() bool {
	return m.Type > All && len(m.After) > 0
}

type ModeType uint8

const (
	UnknownMode ModeType = iota
	All
	AfterTime
	AfterChunk
)

var ModeTypes = map[string]ModeType{
	"all":         All,
	"after-time":  AfterTime,
	"after-chunk": AfterChunk,
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
