package readers

import (
	"log"
	"testing"

	"github.com/elireisman/maven-index-reader-go/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestSimpleIndex(t *testing.T) {
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

	chunkNamesQueue := make(chan string, 1)
	simple := NewIndex(logger, chunkNamesQueue, simpleCfg)
	require.NoError(t, simple.Read())

	chunkCount := 0
	for chunkName := range chunkNamesQueue {
		require.Equal(t, "testdata/nexus-maven-repository-index.gz", chunkName)
		chunkCount++
	}
	require.Equal(t, 1, chunkCount)
}
