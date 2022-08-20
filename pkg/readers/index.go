package readers

import (
	"compress/gzip"
	"log"
	"time"

	"github.com/elireisman/maven-index-reader-go/internal/utils"
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
	l.Printf("Initializing index reader")
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
	ir.logger.Printf("Resolved Nexus timestamp: %s", tsz)

	lastIncr, err := props.GetAsInt("nexus.index.last-incremental")
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}
	ir.logger.Printf("Resolved Nexus last incremented chunk index: %d", lastIncr)

	// validate fetched properties against expected config.Index settings, or bail
	if err := ir.validateProperties(props); err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	targetChunks, err := ir.enumerateIndexChunks(lastIncr)
	if err != nil {
		return errors.Wrap(err, "from Index#Read")
	}

	ir.logger.Printf("Resolved index chunk target list: %v", targetChunks)

	defer close(ir.buffer)
	for _, chunkName := range targetChunks {
		ir.buffer <- chunkName
	}

	return nil
}

// resolve the list of URL or file path suffixes to be
// applied to the base target specified in config.Index
func (ir Index) enumerateIndexChunks(latestChunkID int) ([]string, error) {
	var out []string

	switch ir.cfg.Mode.Type {
	case config.FromChunk:
		// assumption: cfg.Mode.From is the chunk ID of the last
		// successfully processed chunk from the previous run;
		// start consuming for this run from the NEXT chunk ID
		candidateChunkID := int(ir.cfg.Mode.From) + 1

		// TODO: original (Java) version also checks if successor to latestChunkID is already partially available

		// incremental chunk suffix is of the form ".<number>.<file_extension>"
		for candidateChunkID <= latestChunkID {
			candidate := ir.cfg.ResolveTarget(".%d.gz", candidateChunkID)
			if err := ir.remoteChunkExists(candidate); err != nil {
				return out, errors.Wrapf(err, "Index: failed to resolve remote chunk at %s with cause", candidate)
			}

			out = append(out, candidate)
			candidateChunkID++
			time.Sleep(500 * time.Millisecond)
		}

	case config.FromTime:
		// assumption: cfg.Mode.From is a Unix timestamp of the last
		// successfully processed chunk from the previous run
		fromTime := time.UnixMilli(ir.cfg.Mode.From)
		candidateChunkID := latestChunkID

		for candidateChunkID > 0 {
			candidate := ir.cfg.ResolveTarget(".%d.gz", candidateChunkID)
			chunkTime, err := ir.remoteChunkTime(candidate)
			if err != nil {
				return out, errors.Wrapf(err, "Index: failed to obtain timestamp of chunk at %s with cause", candidate)
			}
			if !chunkTime.After(fromTime) {
				break
			}

			out = append(out, candidate)
			candidateChunkID--
			time.Sleep(500 * time.Millisecond)
		}

	default: // config.All
		// full index suffix is of the form ".<file_extension>"
		out = append(out, ir.cfg.ResolveTarget(".gz"))
	}

	return out, nil
}

func (ir Index) remoteChunkExists(target string) error {
	resource, err := resources.FromConfig(ir.logger, ir.cfg, target)
	defer resource.Close()

	_, err = resource.Reader()
	if err != nil {
		return errors.Wrapf(err, "Index: failed to verify resource exists at %s with cause", resource)
	}

	return nil
}

func (ir Index) remoteChunkTime(target string) (time.Time, error) {
	resource, err := resources.FromConfig(ir.logger, ir.cfg, target)
	defer resource.Close()

	rdr, err := resource.Reader()
	if err != nil {
		return time.Now(), errors.Wrapf(err, "Index: failed to verify resource exists at %s with cause", resource)
	}

	gzRdr, err := gzip.NewReader(rdr)
	if err != nil {
		return time.Now(), errors.Wrapf(err, "Index: failed to wrap chunk time check for %s in GZIP Reader with cause", resource)
	}
	defer gzRdr.Close()

	if _, err := utils.ReadByte(gzRdr); err != nil {
		return time.Now(), errors.Wrapf(err, "Index: failed to read chunk %s version with cause", target)
	}

	unixMillis, err := utils.ReadInt64(gzRdr)
	if err != nil {
		return time.Now(), errors.Wrapf(err, "Index: failed to read chunk %s timestamp with cause", target)
	}

	secs := unixMillis / 1000
	nanos := (unixMillis % 1000) * 1000000
	return time.Unix(secs, nanos), nil
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
