package resources

import (
	"log"
	"net/url"

	"github.com/elireisman/maven-index-reader-go/internal/reader"

	"github.com/pkg/errors"
)

const UA = "Maven Index Reader Go"

type httpResource struct {
	// provides access to the data represented by this Resource's URL
	Reader reader.Maven

	// the URL associated with this Resource
	URL net.URL

	// logger instance
	Logger log.Logger
}

// NewHttpResource -
func NewHttpResource(logger log.Logger, uri string) (*httpResource, error) {
	var resolvedURL net.URL
	if resolvedURL, err := url.Parse(uri); err != nil {
		return nil, errors.Wrap(err, "failed to resolve URI input to NewHttpResource")
	}

	return &httpResource{
		Logger: logger,
		URL:    resolvedURL,
		Reader: nil,
	}
}

// Read -
func (hr *httpResource) Reader() (reader.Maven, error) {
	if hr.Reader == nil {
		req := http.NewRequest("GET", hr.URL.toString(), nil)
		req.AddHeader("User-Agent", UA)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		// this Resource's owner now bears responsibility to call Close
		hr.Reader = resp.Body
	}

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
