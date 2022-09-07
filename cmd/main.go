package main

import (
	"flag"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/data/types/record/keys"
	"github.com/elireisman/maven-index-reader-go/pkg/output"
	"github.com/elireisman/maven-index-reader-go/pkg/readers"

	"github.com/pkg/errors"
)

var (
	Format  string
	Out     string
	After   string
	Only    string
	Mode    string
	Pool    int
	Verbose bool
)

func init() {
	flag.StringVar(&Format, "format", "log", "output format: one of 'log', 'json', 'csv'")
	flag.StringVar(&Out, "out", "", "if set, specifies the output file path. stdout if unset")
	flag.StringVar(&After, "after", "", "value depends on --mode; RFC 3339 time string, or int chunk ID, of the last successfully processed chunk")
	flag.StringVar(&Only, "only", "", "value depends on --mode, incompatible with --after; the single chunk ID to process")
	flag.StringVar(&Mode, "mode", "all", "one of 'all', 'after-time', 'after-chunk', 'only-chunk'")
	flag.IntVar(&Pool, "pool", 4, "number of goroutines enabled to scan index chunks in parallel")
	flag.BoolVar(&Verbose, "verbose", false, "log config, skipped records, and progress verbosely")
}

// implements readers.FilterFunc contract to filter
// extracted Maven Central records of interest
func filterFn(record data.Record) bool {
	// skip all but ARTIFACT_ADD and ARTIFACT_REMOVE records
	if record.Type() != data.ArtifactAdd && record.Type() != data.ArtifactRemove {
		return false
	}

	// skip records that aren't associated with a binary artifact,
	// like "javadoc" and "sources" entries
	if record.Get(keys.Classifier) != nil {
		return false
	}

	return true
}

func main() {
	flag.Parse()

	logger := log.Default()

	mavenCentralCfg := config.Index{
		Verbose: Verbose,
		Meta: config.Meta{
			// from https://repo1.maven.org/maven2/.index/nexus-maven-repository-index.properties
			ID:      "central",
			ChainID: "1318453614498",
			File:    "nexus-maven-repository-index",
		},
		Source: config.Source{
			Base: "https://repo1.maven.org/maven2/.index/",
			Type: config.HTTP,
		},
		Mode: config.Mode{
			Type:  config.ModeTypes[strings.ToLower(Mode)],
			After: After,
			Only:  Only,
		},
		Output: config.Output{
			Format: config.OutputFormats[strings.ToLower(Format)],
			File:   Out,
		},
	}
	if err := config.Validate(logger, mavenCentralCfg); err != nil {
		panic(err.Error())
	}

	// Fetch index properties and enumerate index chunks to be scanned
	chunkNamesQueue := make(chan string, 16)
	mavenCentral := readers.NewIndex(logger, chunkNamesQueue, mavenCentralCfg)

	err := mavenCentral.Read()
	if err != nil {
		panic(err.Error())
	}

	// make a queue to buffer records scanned from
	// the various index chunks, and pass it to an
	// output formatter according to CLI args
	records := make(chan data.Record, 64)
	out := output.ResolveFormat(logger, records, mavenCentralCfg)

	// establish a fixed-size worker pool and feed resolved
	// chunks to be scanned into the pool
	var wg sync.WaitGroup
	chunkWorkerPool := make(chan struct{}, Pool)
	for chunkName := range chunkNamesQueue {
		target := chunkName
		wg.Add(1)

		go func() {
			defer func() {
				<-chunkWorkerPool
				wg.Done()
			}()

			chunkWorkerPool <- struct{}{}
			chunk := readers.NewChunk(logger, records, mavenCentralCfg, target, filterFn)
			if err := chunk.Read(); err != nil {
				if errors.Cause(err) == io.EOF {
					logger.Printf("Chunk: EOF encountered for chunk: %s", target)
					return
				}
				logger.Panicf(err.Error())
			}
		}()
	}

	// ensure that when all readers.Chunk goroutines are finished
	// publishing data.Records, the output queue is closed. this
	// will trigger the output formatter to complete and clean up.
	go func() {
		wg.Wait()
		close(records)
	}()

	if err := out.Write(); err != nil {
		panic(err.Error())
	}
}
