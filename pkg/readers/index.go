package readers

import (
	"compress/gzip"
	"log"
	"strconv"
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
	case config.AfterChunk:
		// assumption: cfg.Mode.From is the chunk ID of the last
		// successfully processed chunk from the previous run;
		// start consuming for this run from the NEXT chunk ID
		lastSuccessfulChunk, err := strconv.Atoi(ir.cfg.Mode.After)
		if err != nil {
			return out, errors.Wrapf(err, "Index: failed to parse chunk ID %s with cause", ir.cfg.Mode.After)
		}
		candidateChunkID := lastSuccessfulChunk + 1

		// TODO: original (Java) version also checks if successor to latestChunkID is already partially available

		// incremental chunk suffix is of the form ".<number>.<file_extension>"
		for candidateChunkID <= latestChunkID {
			candidate := ir.cfg.ResolveTarget(".%d.gz", candidateChunkID)
			if err := ir.remoteChunkExists(candidate); err != nil {
				return out, errors.Wrapf(err, "Index: failed to resolve remote chunk at %s with cause", candidate)
			}

			if ir.cfg.Verbose {
				ir.logger.Printf("Index: selected chunk %s", candidate)
			}

			out = append(out, candidate)
			candidateChunkID++
			time.Sleep(500 * time.Millisecond)
		}

	case config.AfterTime:
		// assumption: cfg.Mode.After is a valid RFC 3339 time string of
		// the last successfully processed chunk from the previous run
		fromTime, err := time.Parse(time.RFC3339, ir.cfg.Mode.After)
		if err != nil {
			return out, errors.Wrapf(err, "Index: failed to parse chunk timestamp %s with cause", ir.cfg.Mode.After)
		}
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

			if ir.cfg.Verbose {
				ir.logger.Printf("Index: selected chunk %s with timestamp %s", candidate, chunkTime)
			}

			out = append(out, candidate)
			candidateChunkID--
			time.Sleep(500 * time.Millisecond)
		}

	case config.OnlyChunk:
		chunkID, err := strconv.Atoi(ir.cfg.Mode.Only)
		if err != nil {
			return out, errors.Wrapf(err, "Index: failed to parse chunk ID %s with cause", ir.cfg.Mode.Only)
		}

		candidate := ir.cfg.ResolveTarget(".%d.gz", chunkID)
		if err := ir.remoteChunkExists(candidate); err != nil {
			return out, errors.Wrapf(err, "Index: failed to resolve remote chunk at %s with cause", candidate)
		}

		if ir.cfg.Verbose {
			ir.logger.Printf("Index: selected chunk %s", candidate)
		}
		out = append(out, candidate)

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
	errTime := time.Now().UTC()

	resource, err := resources.FromConfig(ir.logger, ir.cfg, target)
	defer resource.Close()

	rdr, err := resource.Reader()
	if err != nil {
		return errTime, errors.Wrapf(err, "Index: failed to verify resource exists at %s with cause", resource)
	}

	gzRdr, err := gzip.NewReader(rdr)
	if err != nil {
		return errTime, errors.Wrapf(err, "Index: failed to wrap chunk time check for %s in GZIP Reader with cause", resource)
	}
	defer gzRdr.Close()

	if _, err := utils.ReadByte(gzRdr); err != nil {
		return errTime, errors.Wrapf(err, "Index: failed to read chunk %s version with cause", target)
	}

	unixMillis, err := utils.ReadInt64(gzRdr)
	if err != nil {
		return errTime, errors.Wrapf(err, "Index: failed to read chunk %s timestamp with cause", target)
	}

	return time.UnixMilli(unixMillis).UTC(), nil
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
