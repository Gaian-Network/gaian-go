package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetExchangeRatesRequest filters the rate lookup. Both fields are
// optional: an empty Country returns every supported country, and an
// empty SettlementCurrency defaults to "USDC" on the server.
type GetExchangeRatesRequest struct {
	Country            string `json:"country,omitempty"`
	SettlementCurrency string `json:"settlementCurrency,omitempty"`
}

// GetExchangeRatesResponse lists indicative rates. IndicativeOnly is
// always true: only a quote binds a rate. A country missing from Rates
// cannot be quoted right now.
type GetExchangeRatesResponse struct {
	IndicativeOnly bool           `json:"indicativeOnly"`
	Rates          []ExchangeRate `json:"rates"`
}

// ExchangeRate is the indicative rate for one country. ExchangeRate is a
// decimal string with four decimals: fiat units per one settlement token.
type ExchangeRate struct {
	Country            string `json:"country"`
	FiatCurrency       string `json:"fiatCurrency"`
	SettlementCurrency string `json:"settlementCurrency"`
	RouteID            int    `json:"routeId"`
	ExchangeRate       string `json:"exchangeRate"`
	CapturedAt         string `json:"capturedAt"`
}

// GetExchangeRates returns indicative exchange rates without creating a
// quote. The rate a payment settles at is the one Quote returns, which
// may differ.
//
// Example:
//
//	resp, err := client.GetExchangeRates(ctx, &gaian.GetExchangeRatesRequest{Country: "VN"})
//	if err != nil {
//		return err
//	}
//	for _, r := range resp.Data.Rates {
//		fmt.Println(r.Country, r.FiatCurrency, r.ExchangeRate)
//	}
func (c *Client) GetExchangeRates(ctx context.Context, req *GetExchangeRatesRequest) (*PaymentResponse[GetExchangeRatesResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/exchange-rate",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[GetExchangeRatesResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
