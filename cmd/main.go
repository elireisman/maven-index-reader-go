package main

import (
	"flag"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/readers"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"
)

const PropsURL = "https://repo1.maven.org/maven2/.index/nexus-maven-repository-index.properties"

var (
	OutputFormat string
	Target       string
	Backfill     bool
	FromTime     string
)

func init() {
	flag.StringVar(&OutputFormat, "out", "log", "one of 'log', 'json', 'csv'")
	flag.StringVar(&Target, "target", "", "if set, specifies the target output file or path. depends on argument to --out")
	flag.BoolVar(&Backfill, "backfill", false, "if set, indicates a full import of all segments of the source index should be performed")
	flag.StringVar(&FromTime, "from", "", "if set, represents the timestamp of the previous incremental update as a time.Parse-compatible string")
}

func main() {
	flag.Parse()

	logger := log.Default()

	rsc, err := resources.NewHttpResource(logger, PropsURL)
	if err != nil {
		panic(err.Error())
	}

	rdr, err := readers.NewPropertiesReader(logger, rsc)
	if err != nil {
		panic(err.Error())
	}

	props, err := rdr.Read()
	if err != nil {
		panic(err.Error())
	}

	// test timestamp parsing
	tsz, err := props.GetAsTimestamp("nexus.index.timestamp")
	if err != nil {
		panic(err.Error())
	}
	logger.Printf("\nNEXUS TIMESTAMP: %s\n", tsz)
}
