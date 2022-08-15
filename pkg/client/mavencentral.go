package client

import (
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/readers"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

const (
	BaseURL = "https://repo1.maven.org/maven2/.index/"

	PropertiesFile = "nexus-maven-repository-index.properties"

	FullIndexFile = "nexus-maven-repository-index.gz"

	IncrementalFilePattern = "nexus-maven-repository-index.%d.gz"
)

type MavenCentral struct {
	cfg config.Index

	logger *log.Logger

	out chan data.Record
}

func NewMavenCentral(l *log.Logger, out chan data.Record, cfg config.Index) MavenCentral {
	return MavenCentral{
		cfg:    cfg,
		logger: l,
		out:    out,
	}
}

// Start -
func (mc MavenCentral) Start() error {
	// 1. load props file
	propsURL := BaseURL + PropertiesFile
	rsc, err := resources.NewHttpResource(mc.logger, propsURL)
	if err != nil {
		return errors.Wrap(err, "in MavenCentral#Start")
	}

	rdr, err := readers.NewPropertiesReader(mc.logger, rsc)
	if err != nil {
		return errors.Wrap(err, "in MavenCentral#Start")
	}

	props, err := rdr.Read()
	if err != nil {
		return errors.Wrap(err, "in MavenCentral#Start")
	}

	// parse out important properties values
	tsz, err := props.GetAsTimestamp("nexus.index.timestamp")
	if err != nil {
		return errors.Wrap(err, "in MavenCentral#Start")
	}
	mc.logger.Printf("Resolved Nexus timestamp: %s\n", tsz)

	lastIncr, err := props.GetAsInt("nexus.index.last-incremental")
	if err != nil {
		return errors.Wrap(err, "in MavenCentral#Start")
	}
	mc.logger.Printf("Resolved Nexus last incremented chunk index: %d", lastIncr)

	// 2. compare to config.Index settings
	// 3. build IndexReader (which will build and regulate with queue + fixed worker pool for ChunkReaders)
	// 4. caller will consume from given "chan data.Record" into selected output format!

	// TODO(eli): IMPLEMENT!
	return nil
}
