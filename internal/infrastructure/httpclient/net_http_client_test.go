package httpclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mi7teR/exr/mocks"
	"go.uber.org/mock/gomock"
)

func TestNewNetHTTPClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	client := NewNetHTTPClient(mockLogger)

	// Assert that client is not nil
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	httpClient := client
	// Assert timeout is set correctly
	if httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout to be %v, got %v", DefaultTimeout, httpClient.Timeout)
	}

	// Assert Transport is of type *httpclient.LogRoundTripper
	_, ok := httpClient.Transport.(*LogRoundTripper)
	if !ok {
		t.Errorf("Expected Transport to be *httpclient.LogRoundTripper, got %T", httpClient.Transport)
	}
}

func TestLogRoundTripper_RoundTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)

	// Set up a test server to handle the HTTP request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create the LogRoundTripper with a default transport
	originalTransport := http.DefaultTransport
	lrt := &LogRoundTripper{
		l:         mockLogger,
		transport: originalTransport,
	}

	// Set up expectations on the mock logger
	mockLogger.EXPECT().Info(
		"http request completed",
		"method", "GET",
		"url", gomock.Any(),
		"status", "200 OK",
		"duration", gomock.Any(),
	)

	// Create an HTTP client with the LogRoundTripper
	client := &http.Client{
		Transport: lrt,
	}

	// Perform the HTTP request
	_, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// error transport returning only error (resp == nil)
type errorTransport struct{ err error }

func (e *errorTransport) RoundTrip(_ *http.Request) (*http.Response, error) { return nil, e.err }

// error transport returning resp + error
type errorWithRespTransport struct{ err error }

func (e *errorWithRespTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("fail")),
		Request:    req,
	}, e.err
}

func TestLogRoundTripper_RoundTrip_Error_NoResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	testErr := errors.New("boom")

	lrt := &LogRoundTripper{
		l:         mockLogger,
		transport: &errorTransport{err: testErr},
	}

	mockLogger.EXPECT().Error(
		"http request failed",
		"method", "GET",
		"url", gomock.Any(),
		"status", "",
		"error", testErr,
		"duration", gomock.Any(),
	)

	client := &http.Client{Transport: lrt}
	_, err := client.Get("http://example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestLogRoundTripper_RoundTrip_Error_WithResp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	testErr := errors.New("fail")

	lrt := &LogRoundTripper{
		l:         mockLogger,
		transport: &errorWithRespTransport{err: testErr},
	}

	mockLogger.EXPECT().Error(
		"http request failed",
		"method", "GET",
		"url", gomock.Any(),
		"status", "500 Internal Server Error",
		"error", testErr,
		"duration", gomock.Any(),
	)

	client := &http.Client{Transport: lrt}
	_, err := client.Get("http://example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
