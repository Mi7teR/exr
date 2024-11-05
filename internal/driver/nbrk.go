package driver

import (
	"context"
	"encoding/xml"
	"net/http"

	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/errors"
)

type NBRK struct {
	addr       string
	httpClient HTTPClient
}

// NewNBRK creates a new NBRK driver.
// default address is "https://nationalbank.kz/rss/rates_all.xml"
// but you can pass your own address.
func NewNBRK(addr string, client HTTPClient) *NBRK {
	return &NBRK{
		addr:       addr,
		httpClient: client,
	}
}

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Item []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			PubDate     string `xml:"pubDate"`
			Description string `xml:"description"`
			Quant       string `xml:"quant"`
			Index       string `xml:"index"`
			Change      string `xml:"change"`
			Link        string `xml:"link"`
		} `xml:"item"`
	} `xml:"channel"`
}

// FetchRates fetches exchange rates from the NBRK API.
func (n *NBRK) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, n.addr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Parse the response body.
	var rssData rss
	if err = xml.NewDecoder(resp.Body).Decode(&rssData); err != nil {
		return nil, err
	}

	// Convert the response to a list of exchange rates.
	var rates []*entity.ExchangeRate
	for _, item := range rssData.Channel.Item {
		if !canPerformCurrency(item.Title) {
			continue
		}
		rate := &entity.ExchangeRate{
			Source:       "NBRK",
			CurrencyCode: item.Title,
			Buy:          item.Description,
			Sell:         item.Description,
		}

		rates = append(rates, rate)
	}

	if len(rates) == 0 {
		return nil, errors.ErrNotFound
	}

	return rates, nil
}

// canPerformCurrency checks if the currency is supported by exr api.
func canPerformCurrency(currency string) bool {
	switch currency {
	case "USD", "EUR", "RUB":
		return true
	}
	return false
}
