package httpclient

import (
	"net/http"
	"net/http/httptest"
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
