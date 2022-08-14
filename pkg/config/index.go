package config

import "time"

// configuration for an readers.IndexReader
type Index struct {
	Name        string
	Incremental bool
	FromChunk   uint
	FromTime    time.Time
	BaseURL     string
}
