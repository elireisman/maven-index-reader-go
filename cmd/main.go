package main

import (
	"fmt"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/readers"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"
)

const PropsURL = "https://repo1.maven.org/maven2/.index/nexus-maven-repository-index.properties"

func main() {
	logger := log.Default()

	rsc, err := resources.NewHttpResource(logger, PropsURL)
	if err != nil {
		panic(err.Error())
	}

	rdr, err := readers.NewProperties(logger, rsc)
	if err != nil {
		panic(err.Error())
	}

	if err = rdr.Execute(); err != nil {
		panic(err.Error())
	}

	// test timestamp parsing
	tsz, err := rdr.GetAsTimestamp("nexus.index.timestamp")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("\nNEXUS TIMESTAMP: %s\n", tsz)
}
