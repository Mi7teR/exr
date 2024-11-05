package driver

import (
	"context"
	"encoding/xml"

	"github.com/Mi7teR/exr/internal/entity"
)

type NBRK struct {
	addr       string
	httpClient HTTPClient
}

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
	// Fetch exchange rates from the NBRK API.
	// ...
	return nil, nil
}
