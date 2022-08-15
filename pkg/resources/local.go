package resources

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
)

type localResource struct {
	Logger *log.Logger

	// path to a local file to open
	Path string

	reader io.ReadCloser
}

func NewLocalResource(l *log.Logger, path string) (*localResource, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.IsDir() {
		return nil, errors.Wrapf(err, "NewLocalResource: failed to stat expected file at %s with cause", path)
	}

	return &localResource{
		Logger: l,
		Path:   path,
		reader: nil,
	}, nil
}

func (lr localResource) Reader() (io.Reader, error) {
	if lr.reader != nil {
		return nil, errors.Errorf("LocalResource(%s): unexpected Reader() call on non-nil io.ReadCloser", lr.Path)
	}

	f, err := os.Open(lr.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "LocalResource: failed to open expected file at %s with cause", lr.Path)
	}

	lr.reader = f
	buf := bufio.NewReader(f)

	return buf, nil
}

func (lr localResource) Close() error {
	if lr.reader == nil {
		return errors.Errorf("LocalResource(%s): unexpected Close() call on nil io.ReadCloser", lr.Path)
	}

	return lr.reader.Close()
}
