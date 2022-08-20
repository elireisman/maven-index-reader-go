package main

import (
	"flag"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/output"
	"github.com/elireisman/maven-index-reader-go/pkg/readers"

	"github.com/pkg/errors"
)

var (
	Format string
	Out    string
	From   int64
	Mode   string
	Pool   int
)

func init() {
	flag.StringVar(&Format, "format", "log", "output format: one of 'log', 'json', 'csv'")
	flag.StringVar(&Out, "out", "", "if set, specifies the output file path. stdout if unset")
	flag.Int64Var(&From, "from", 0, "if non-zero, specifies the chunk timestamp as Unix milliseconds, "+
		"or oldest chunk ID to process, in an incremental run. depends on --mode")
	flag.StringVar(&Mode, "mode", "all", "one of 'all', 'from-time', 'from-chunk'")
	flag.IntVar(&Pool, "pool", 4, "number of goroutines enabled to scan index chunks in parallel")
}

func main() {
	flag.Parse()

	logger := log.Default()

	mavenCentralCfg := config.Index{
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
			Type: config.ModeTypes[strings.ToLower(Mode)],
			From: From,
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
			chunk := readers.NewChunk(logger, records, mavenCentralCfg, target)
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
