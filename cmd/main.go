package main

import (
	"flag"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	//"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/readers"
)

var (
	OutMode     string
	OutFile     string
	Incremental bool
	From        int
)

func init() {
	flag.StringVar(&OutMode, "out-mode", "log", "one of 'log', 'json', 'csv'")
	flag.StringVar(&OutFile, "out-file", "", "if set, specifies the target output file or path. Depends on --out-mode")
	flag.BoolVar(&Incremental, "incremental", false, "if set, perform an incremental read only, rather than full index")
	flag.IntVar(&From, "from", 0, "update from the provided chunk ID to most recent, only. Depends on --incremental")
}

func main() {
	flag.Parse()

	logger := log.Default()
	out := make(chan readers.Chunk, 16)

	mavenCentralCfg := config.Index{
		Meta: config.Meta{
			ID:      "central",
			ChainID: 1318453614498, // from https://repo1.maven.org/maven2/.index/nexus-maven-repository-index.properties
			Target:  "nexus-maven-repository-index",
		},
		Source: config.Source{
			Base: "https://repo1.maven.org/maven2/.index/",
			Type: config.HTTP,
		},
		Mode: config.Mode{
			Incremental: true,
			FromChunk:   767,
		},
	}

	mavenCentral := readers.NewIndex(logger, out, mavenCentralCfg)

	err := mavenCentral.Read()
	if err != nil {
		panic(err.Error())
	}
}
