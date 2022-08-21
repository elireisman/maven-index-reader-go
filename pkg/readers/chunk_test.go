package readers

import (
	"io"
	"log"
	"testing"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSimpleChunk(t *testing.T) {
	logger := log.Default()

	simpleCfg := config.Index{
		Meta: config.Meta{
			ID:      "apache-snapshots-local",
			ChainID: "1243533418968",
			File:    "nexus-maven-repository-index",
		},
		Source: config.Source{
			Base: "testdata/",
			Type: config.Local,
		},
		Mode: config.Mode{
			Type: config.All,
		},
		Output: config.Output{
			Format: config.Log,
		},
	}
	require.NoError(t, config.Validate(logger, simpleCfg))

	target := simpleCfg.ResolveTarget(".gz")

	records := make(chan data.Record, 5)
	chunk := NewChunk(logger, records, simpleCfg, target)

	err := chunk.Read()
	require.True(t, errors.Cause(err) == io.EOF, "(%T) %s", err, err)

	record := <-records
	close(records)

	require.Equal(t, record.Type(), data.ArtifactAdd)
	require.Equal(t, "org.sonatype.nexus", record.Get("groupId"))
	require.Equal(t, "nexus", record.Get("artifactId"))
	require.Equal(t, "1.3.0-SNAPSHOT", record.Get("version"))
	require.Equal(t, "Nexus Repository Manager", record.Get("name"))
	require.Equal(t, "pom", record.Get("packaging"))
	require.Equal(t, "pom", record.Get("fileExtension"))
	require.Equal(t, int64(1243533415343), record.Get("fileModified"))
}
