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
		// no Filter clause - allow all RecordTypes for this
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
	require.Equal(t, data.ArtifactAdd, record.Type())
	require.Equal(t, "org.sonatype.nexus", record.Get("groupId"))
	require.Equal(t, "nexus", record.Get("artifactId"))
	require.Equal(t, "1.3.0-SNAPSHOT", record.Get("version"))
	require.Equal(t, "Nexus Repository Manager", record.Get("name"))
	require.Equal(t, "pom", record.Get("packaging"))
	require.Equal(t, "pom", record.Get("fileExtension"))
	require.Equal(t, int64(1243533415343), record.Get("fileModified"))

	record = <-records
	require.Equal(t, data.ArtifactAdd, record.Type())
	require.Equal(t, "org.sonatype.test-evict", record.Get("groupId"))
	require.Equal(t, "sonatype-test-evict_1.4_mail", record.Get("artifactId"))
	require.Equal(t, "1.0-SNAPSHOT", record.Get("version"))
	require.Equal(t, "jar", record.Get("packaging"))
	require.Equal(t, "jar", record.Get("fileExtension"))
	require.Equal(t, int64(1243533415359), record.Get("fileModified"))

	record = <-records
	require.Equal(t, data.RootGroups, record.Type())
	require.Equal(t, []string{"org"}, record.Get("rootGroupsList"))

	record = <-records
	require.Equal(t, data.AllGroups, record.Type())
	require.ElementsMatch(t, []string{"org.sonatype.test-evict", "org.sonatype.nexus"}, record.Get("allGroupsList"))

	record = <-records
	require.Equal(t, data.Descriptor, record.Type())
	require.Equal(t, "apache-snapshots", record.Get("repositoryId"))
	require.Equal(t, "1.0", record.Get("version"))
	require.Equal(t, "NexusIndex", record.Get("DESCRIPTOR"))

	close(records)
}
