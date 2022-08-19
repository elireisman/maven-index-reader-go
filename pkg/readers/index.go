package readers

import (
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
	target := ir.cfg.ResolveTarget(".properties")
	rsc, err := resources.FromConfig(ir.logger, ir.cfg, target)
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

	targetChunks := ir.enumerateIndexChunks(lastIncr)
	ir.logger.Printf("Resolved index chunk target list: %v", targetChunks)

	defer close(ir.buffer)
	for _, chunkName := range targetChunks {
		ir.buffer <- chunkName
	}

	return nil
}

// resolve the list of URL or file path suffixes to be
// applied to the base target specified in config.Index
func (ir Index) enumerateIndexChunks(latestChunk int) []string {
	var out []string

	// TODO(eli): check that NEXT (+1) CHUNK ISN'T ALSO AVAILABLE? USE HTTP HEAD REQS TO CHECK? OR TSZ + VER SCAN PER-CHUNK?
	if ir.cfg.Mode.IsIncrementalRun() {
		// assumption: this value was incremented at the end of the last
		// successful run, and has not been successfully consumed yet
		prevChunk := ir.cfg.Mode.FromChunk

		// incremental chunk suffix is of the form ".<number>.<file_extension>"
		for prevChunk <= latestChunk {
			out = append(out, ir.cfg.ResolveTarget(".%d.gz", prevChunk))
			prevChunk++
		}
	} else {
		// full index suffix is of the form ".<file_extension>"
		out = append(out, ir.cfg.ResolveTarget(".gz"))
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
