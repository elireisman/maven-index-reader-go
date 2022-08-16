package readers

import (
	"fmt"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type Index struct {
	cfg    config.Index
	logger *log.Logger
	buffer chan<- string
}

func NewIndex(l *log.Logger, b chan<- string, c config.Index) Index {
	l.Printf("Initializing Index reader with configuration: %+v", c)

	return Index{
		cfg:    c,
		logger: l,
		buffer: b,
	}
}

func (ir Index) Read() error {
	// load remote index properties file
	rsc, err := resources.ConfigureResource(ir.logger, ir.cfg, ".properties")
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	rdr, err := NewProperties(ir.logger, rsc)
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	props, err := rdr.Read()
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	// parse out important properties values
	tsz, err := props.GetAsTimestamp("nexus.index.timestamp")
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}
	ir.logger.Printf("Resolved Nexus timestamp: %s\n", tsz)

	lastIncr, err := props.GetAsInt("nexus.index.last-incremental")
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}
	ir.logger.Printf("Resolved Nexus last incremented chunk index: %d", lastIncr)

	// validate fetched properties against expected config.Index settings, or bail
	if err := ir.validateProperties(props); err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	targetChunks := ir.createChunkSuffixList(lastIncr)
	ir.logger.Printf("Resolved chunk suffix list: %v", targetChunks)

	defer close(ir.buffer)
	for _, chunkName := range targetChunks {
		ir.buffer <- chunkName
	}

	return nil
}

// resolve the list of URL or file path suffixes to be
// applied to the base target specified in config.Index
func (ir Index) createChunkSuffixList(latestChunk int) []string {
	var out []string

	// TODO(eli): check that NEXT (+1) CHUNK ISN'T ALSO AVAILABLE? USE HTTP HEAD REQS TO CHECK? OR TSZ + VER SCAN PER-CHUNK?
	if ir.cfg.Mode.Incremental {
		// assumption: this was incremented during LAST SUCCESSFUL RUN and is UNSEEN as of now!
		prevChunk := ir.cfg.Mode.FromChunk
		// incremental chunk suffix is of the form ".<number>.<file_extension>"
		incrementalChunkSuffixPattern := ".%d.gz"
		for prevChunk <= latestChunk {
			out = append(out, fmt.Sprintf(incrementalChunkSuffixPattern, prevChunk))
			prevChunk++
		}
	} else {
		// full index suffix is of the form ".<file_extension>"
		fullIndexChunkSuffix := ".gz"
		out = append(out, fullIndexChunkSuffix)
	}

	return out
}

func (ir Index) validateProperties(props data.Properties) error {
	indexID, err := props.GetAsString("nexus.index.id")
	if err != nil {
		return err
	}
	if ir.cfg.Meta.ID != indexID {
		return errors.Errorf("failed to validate expected index ID %s, got: %s", ir.cfg.Meta.ID, indexID)
	}

	chainID, err := props.GetAsString("nexus.index.chain-id")
	if err != nil {
		return err
	}
	if ir.cfg.Meta.ChainID != chainID {
		return errors.Errorf("failed to validate expected chain ID %s, got: %s", ir.cfg.Meta.ChainID, chainID)
	}

	return nil
}
