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
	Reader io.ReadCloser

	// the URL associated with this Resource
	URL net.URL

	// logger instance
	Logger log.Logger
}

// NewHttpResource -
func NewHttpResource(logger log.Logger, uri string) (*httpResource, error) {
	if _, err := url.Parse(uri); err != nil {
		return nil, errors.Wrap(err, "NewHttpResource: invalid URI %q with cause", uri)
	}

	return &httpResource{
		Logger: logger,
		URL:    uri,
		Reader: nil,
	}
}

// Read -
func (hr *httpResource) Reader() (io.Reader, error) {
	if hr.Reader != nil {
		return nil, errors.Error("HttpResource: unexpected repeat call to Reader()")
	}

	req := http.NewRequest("GET", hr.URL, nil)
	req.AddHeader("User-Agent", UA)
	req.AddHeader("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// this Resource's owner now bears responsibility to call Close
	hr.Reader = resp.Body // TODO(eli): do we need explicit gzip.NewReader(body) here?
	return hr.Reader, nil
}

// Close -
func (hr *httpResource) Close() error {
	if hr.Reader != nil {
		hr.Reader.Close()
		return nil
	}

	return errors.New("unexpected Close call on HttpResource's nil Reader")
}
