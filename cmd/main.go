package main

import (
	"fmt"
	"log"

	"github.com/elireisman/maven-index-reader-go/internal/util"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"
)

const PropsURL = "https://repo1.maven.org/maven2/.index/nexus-maven-repository-index.properties"

func main() {
	rsc, err := resources.NewHttpResource(log.Default(), PropsURL)
	if err != nil {
		panic(err.Error())
	}

	rdr, err := rsc.Reader()
	if err != nil {
		panic(err.Error())
	}

	props, err := util.GetProperties(rdr)
	if err != nil {
		panic(err.Error())
	}

	for k, v := range props {
		fmt.Printf("KEY(%s) => VALUE(%s)\n", k, v)
	}
}
