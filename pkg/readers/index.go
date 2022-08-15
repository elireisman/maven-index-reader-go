package readers

import (
	"log"

	//"github.com/elireisman/maven-index-reader-go/internal/utils"
	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
	"github.com/elireisman/maven-index-reader-go/pkg/resources"

	"github.com/pkg/errors"
)

type Index struct {
	cfg    config.Index
	logger *log.Logger
	buffer chan Chunk
}

func NewIndex(l *log.Logger, b chan Chunk, c config.Index) Index {
	l.Printf("Initializing Index reader with configuration: %+v", c)

	return Index{
		cfg:    c,
		logger: l,
		buffer: b,
	}
}

func (ir Index) Read() error {
	// load remote index properties file
	rsc, err := ir.resourceFromConfig(".properties")
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

	// TODO(eli): IMPLEMENT!

	return nil
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

// resolve Resource location from config.Index supplied to readers.Index
func (ir Index) resourceFromConfig(suffix string) (resources.Resource, error) {
	var resource resources.Resource
	var err error

	// this can be a URL to a remote resource or a local file path
	target := ir.cfg.Source.Base + ir.cfg.Meta.Target + suffix

	switch ir.cfg.Source.Type {
	case config.Local:
		resource, err = resources.NewLocalResource(ir.logger, target)
	case config.HTTP:
		resource, err = resources.NewHttpResource(ir.logger, target)
	default:
		err = errors.Errorf("Index: invalid config.Index.Source.Type for target %s, got: %d", target, ir.cfg.Source.Type)
	}

	return resource, err
}
