package main

import (
	"flag"
	"io"
	"log"
	"sync"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/output"
	"github.com/elireisman/maven-index-reader-go/pkg/readers"

	"github.com/pkg/errors"
)

var (
	OutMode     string
	OutFile     string
	Incremental bool
	From        int
	Concurrency int
)

func init() {
	flag.StringVar(&OutMode, "out-mode", "log", "one of 'log', 'json', 'csv'")
	flag.StringVar(&OutFile, "out-file", "", "if set, specifies the target output file or path. Depends on --out-mode")
	flag.IntVar(&From, "from", 0, "if non-zero, only process index chunk updates from the provided chunk ID to most recent")
	flag.IntVar(&Concurrency, "concurrency", 4, "number of goroutines enabled to scan index chunks in parallel")
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
			FromChunk: From,
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
	outputQueue := make(chan data.Record, 64)
	var out output.Output
	switch OutMode {
	case "json":
		out = output.NewJSON(logger, outputQueue, OutFile)
	case "csv":
		out = output.NewCSV(logger, outputQueue, OutFile)
	default:
		// TODO(eli): eliminate this - redundant! for debug: prints Go structs to stdout
		out = output.NewLogger(logger, outputQueue)
	}

	// TODO(eli): MOVE MOST OF THE BELOW INTO readers.Index ?!?

	// set up a fixed-size worker pool and feed resolved
	// chunks to be scanned into the pool
	var wg sync.WaitGroup
	chunkWorkerPool := make(chan struct{}, Concurrency)
	for chunkName := range chunkNamesQueue {
		target := chunkName
		wg.Add(1)

		chunkWorkerPool <- struct{}{}
		go func() {
			defer func() {
				<-chunkWorkerPool
				wg.Done()
			}()

			chunk := readers.NewChunk(logger, outputQueue, mavenCentralCfg, target)
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
		close(outputQueue)
	}()

	if err := out.Write(); err != nil {
		panic(err.Error())
	}
}
