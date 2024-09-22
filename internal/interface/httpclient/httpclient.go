package httpclient

import "net/http"

// HTTPClient is an interface that defines the methods that an HTTP client must implement.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
