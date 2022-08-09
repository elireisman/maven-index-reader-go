package resources

import (
	"io"
	"log"
	"net/http"
	"net/url"

	//"github.com/elireisman/maven-index-reader-go/internal/util"

	"github.com/pkg/errors"
)

const UA = "Maven Index Reader Go"

type httpResource struct {
	// provides access to the data represented by this Resource's URL
	reader io.ReadCloser

	// the URL associated with this Resource
	URL string

	// logger instance
	Logger *log.Logger
}

// NewHttpResource -
func NewHttpResource(logger *log.Logger, uri string) (*httpResource, error) {
	if _, err := url.Parse(uri); err != nil {
		return nil, errors.Wrapf(err, "NewHttpResource: invalid URI %q with cause", uri)
	}

	return &httpResource{
		Logger: logger,
		URL:    uri,
		reader: nil,
	}, nil
}

// Read -
func (hr *httpResource) Reader() (io.Reader, error) {
	if hr.reader != nil {
		return nil, errors.New("HttpResource: unexpected repeat call to Reader()")
	}

	req, err := http.NewRequest("GET", hr.URL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "HttpResource: failed to build GET req to %s with cause", hr.URL)
	}

	req.Header.Add("User-Agent", UA)
	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// this Resource's owner now bears responsibility to call Close
	hr.reader = resp.Body // TODO(eli): do we need explicit gzip.NewReader(body) here?
	return hr.reader, nil
}

// Close -
func (hr *httpResource) Close() error {
	if hr.reader != nil {
		hr.reader.Close()
		return nil
	}

	return errors.New("unexpected Close call on HttpResource's nil Reader")
}
