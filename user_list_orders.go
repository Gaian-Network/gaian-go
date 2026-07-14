package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListUserOrdersRequest paginates and optionally filters a user's order
// history. Status, Page, and PageSize are all optional — leave them at
// the zero value to get the API's defaults (all statuses, page 1, 20
// per page).
type ListUserOrdersRequest struct {
	UserID   string `param:"userId" json:"-"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// ListUserOrdersResponse is a page of a user's order history.
//
// Unlike every other endpoint in this SDK, RequestID/Items/Pagination
// are siblings at the top level of the response body — there's no outer
// {data: ...} envelope wrapping this one. So, unlike the rest of the
// User-service endpoints, ListUserOrders returns *ListUserOrdersResponse
// directly instead of *UserResposne[ListUserOrdersResponse].
type ListUserOrdersResponse struct {
	RequestID  string      `json:"requestId"`
	Items      []UserOrder `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// UserOrder is the order shape returned by the User service's order-list
// and order-detail endpoints (ListUserOrders / GetUserOrder). It is
// intentionally a distinct type from Order (payment_get_order_status.go,
// the Payment service's GET /orders/{id} shape) — same real-world thing,
// two different backing services with different field sets. Don't mix
// them up.
type UserOrder struct {
	OrderID        string             `json:"orderId"`
	Status         string             `json:"status"`
	FiatAmount     string             `json:"fiatAmount"`
	FiatCurrency   string             `json:"fiatCurrency"`
	CryptoAmount   string             `json:"cryptoAmount"`
	CryptoCurrency string             `json:"cryptoCurrency"`
	ExchangeRate   string             `json:"exchangeRate"`
	ProtocolFee    string             `json:"protocolFee"`
	ExpiresAt      string             `json:"expiresAt"`
	CreatedAt      string             `json:"createdAt"`
	Payment        PaymentInfo        `json:"payment"`
	CryptoTransfer CryptoTransferInfo `json:"cryptoTransfer"`
}

// Pagination carries paging metadata for list endpoints.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
}

// PaymentInfo is the fiat settlement leg of a UserOrder.
type PaymentInfo struct {
	Amount      string  `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Provider    string  `json:"provider"`
	ProviderRef string  `json:"providerRef"`
	PaidAt      *string `json:"paidAt,omitempty"`
}

// CryptoTransferInfo is the on-chain leg of a UserOrder.
type CryptoTransferInfo struct {
	Amount      string  `json:"amount"`
	Chain       string  `json:"chain"`
	Token       string  `json:"token"`
	FromAddress string  `json:"fromAddress"`
	ToAddress   string  `json:"toAddress"`
	TxHash      string  `json:"txHash"`
	Status      string  `json:"status"`
	IsConfirmed bool    `json:"isConfirmed"`
	ConfirmedAt *string `json:"confirmedAt,omitempty"`
}

// ListUserOrders returns a page of order history for req.UserID. Use
// GetUserOrder to fetch a single order by ID instead of paging through
// this list.
//
// Example:
//
//	resp, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{
//		UserID: userID,
//		Page:   1,
//	})
//	if err != nil {
//		return err
//	}
//	for _, o := range resp.Items {
//		fmt.Println(o.OrderID, o.Status, o.FiatAmount, o.FiatCurrency)
//	}
func (c *Client) ListUserOrders(ctx context.Context, req *ListUserOrdersRequest) (*ListUserOrdersResponse, error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/orders", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(ListUserOrdersResponse)
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
