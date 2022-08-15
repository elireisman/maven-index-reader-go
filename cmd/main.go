package main

import (
	"flag"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/client"
	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
)

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
	out := make(chan data.Record, 128)
	cfg := config.Index{}

	mavenCentral := client.NewMavenCentral(logger, out, cfg)

	err := mavenCentral.Start()
	if err != nil {
		panic(err.Error())
	}
}
