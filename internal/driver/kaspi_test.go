package driver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mi7teR/exr/internal/driver"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type kaspiMockResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Body    []struct {
		Currency string `json:"currency"`
		Buy      int    `json:"buy"`
		Sale     int    `json:"sale"`
	} `json:"body"`
}

func TestKaspi_FetchRates(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   kaspiMockResponse
		responseStatus int
		expectedRates  []*entity.ExchangeRate
		expectedError  bool
	}{
		{
			name: "Successful fetch",
			responseBody: kaspiMockResponse{
				Status:  "OK",
				Message: "OK",
				Body: []struct {
					Currency string `json:"currency"`
					Buy      int    `json:"buy"`
					Sale     int    `json:"sale"`
				}{
					{"USD", 450, 460},
					{"EUR", 510, 520},
				},
			},
			responseStatus: http.StatusOK,
			expectedRates: []*entity.ExchangeRate{
				{Source: "Kaspi", CurrencyCode: "USD", Buy: "450", Sell: "460"},
				{Source: "Kaspi", CurrencyCode: "EUR", Buy: "510", Sell: "520"},
			},
			expectedError: false,
		},
		{
			name:           "API returns an error",
			responseBody:   kaspiMockResponse{Status: "ERROR", Message: "Internal error"},
			responseStatus: http.StatusInternalServerError,
			expectedRates:  nil,
			expectedError:  true,
		},
		{
			name:           "Empty response body",
			responseBody:   kaspiMockResponse{},
			responseStatus: http.StatusOK,
			expectedRates:  nil,
			expectedError:  true,
		},
		{
			name:           "Invalid JSON response",
			responseBody:   kaspiMockResponse{},
			responseStatus: http.StatusOK,
			expectedRates:  nil,
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.responseStatus)
				if test.name == "Invalid JSON response" {
					_, _ = w.Write([]byte("{invalid json"))
				} else {
					responseBytes, _ := json.Marshal(test.responseBody)
					_, _ = w.Write(responseBytes)
				}
			}))
			defer server.Close()

			kaspi := driver.NewKaspi(server.URL, server.Client())
			ctx := context.Background()

			rates, err := kaspi.FetchRates(ctx)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedRates, rates)
			}
		})
	}
}
