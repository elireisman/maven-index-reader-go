package main

import (
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
