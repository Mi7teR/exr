package driver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mi7teR/exr/internal/entity"
	internalErrors "github.com/Mi7teR/exr/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_canPerformCurrency(t *testing.T) {
	type args struct {
		currency string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "USD",
			args: args{currency: "USD"},
			want: true,
		},
		{
			name: "EUR",
			args: args{currency: "EUR"},
			want: true,
		},
		{
			name: "RUB",
			args: args{currency: "RUB"},
			want: true,
		},
		{
			name: "unsupported",
			args: args{currency: "unsupported"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canPerformCurrency(tt.args.currency); got != tt.want {
				t.Errorf("canPerformCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchRates(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		expectedRates  []*entity.ExchangeRate
		expectedError  error
	}{
		{
			name: "Successful fetch",
			responseBody: `
				<rss>
					<channel>
						<item>
							<title>USD</title>
							<pubDate>Mon, 01 Nov 2024 12:00:00 +0000</pubDate>
							<description>456.00</description>
						</item>
						<item>
							<title>EUR</title>
							<pubDate>Mon, 01 Nov 2024 12:00:00 +0000</pubDate>
							<description>512.00</description>
						</item>
						<item>
							<title>RUB</title>
							<pubDate>Mon, 01 Nov 2024 12:00:00 +0000</pubDate>
							<description>6.00</description>
						</item>
					</channel>
				</rss>
			`,
			responseStatus: http.StatusOK,
			expectedRates: []*entity.ExchangeRate{
				{
					Source:       "NBRK",
					CurrencyCode: "USD",
					Buy:          "456.00",
					Sell:         "456.00",
				},
				{
					Source:       "NBRK",
					CurrencyCode: "EUR",
					Buy:          "512.00",
					Sell:         "512.00",
				},
				{
					Source:       "NBRK",
					CurrencyCode: "RUB",
					Buy:          "6.00",
					Sell:         "6.00",
				},
			},
			expectedError: nil,
		},
		{
			name:           "Empty response",
			responseBody:   "<rss><channel></channel></rss>",
			responseStatus: http.StatusOK,
			expectedRates:  nil,
			expectedError:  internalErrors.ErrNotFound,
		},
		{
			name:           "Invalid XML",
			responseBody:   "Invalid XML",
			responseStatus: http.StatusOK,
			expectedRates:  nil,
			expectedError:  errors.New("EOF"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.responseStatus)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer server.Close()

			nbrk := NewNBRK(server.URL, server.Client())
			ctx := context.Background()

			rates, err := nbrk.FetchRates(ctx)
			if test.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedRates, rates)
			}
		})
	}
}
