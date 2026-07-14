package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OrderStatus is the lifecycle state of an order, from GetOrderStatus.
type OrderStatus int

const (
	OrderPending         OrderStatus = 0
	OrderAwaitingDeposit OrderStatus = 1
	OrderPaymentReceived OrderStatus = 2
	OrderQueued          OrderStatus = 3
	OrderProcessing      OrderStatus = 4
	OrderCompleted       OrderStatus = 10
	OrderFailed          OrderStatus = 20
	OrderCancelled       OrderStatus = 21
	OrderExpired         OrderStatus = 22
	OrderUnknown         OrderStatus = 99
)

// IsTerminal reports whether the order has reached a final state
// (OrderCompleted, OrderFailed, OrderCancelled, or OrderExpired). Poll
// GetOrderStatus until this returns true.
func (s OrderStatus) IsTerminal() bool {
	switch s {
	case OrderCompleted, OrderFailed, OrderCancelled, OrderExpired:
		return true
	default:
		return false
	}
}

// GetOrderStatusRequest identifies the order to look up.
type GetOrderStatusRequest struct {
	OrderID string `param:"orderId"`
}

// Order is the Payment service's order representation. Distinct from
// UserOrder (user_list_orders.go, the User service's
// ListUserOrders/GetUserOrder shape) — same real-world thing, two
// different backing services with different field sets.
type Order struct {
	OrderID            string      `json:"orderId"`
	Status             OrderStatus `json:"status"`
	StatusLabel        string      `json:"statusLabel"`
	ChainID            ChainID     `json:"chainId"`
	DepositAddress     *string     `json:"depositAddress,omitempty"`
	TransactionHash    *string     `json:"transactionHash,omitempty"`
	FiatAmount         float64     `json:"fiatAmount"`
	FiatCurrency       string      `json:"fiatCurrency"`
	SettlementAmount   string      `json:"settlementAmount"`
	SettlementCurrency string      `json:"settlementCurrency"`
	ExchangeRate       string      `json:"exchangeRate"`
	CreatedAt          string      `json:"createdAt"`
	UpdatedAt          string      `json:"updatedAt"`
}

// GetOrderStatusResponse is the current state of an order.
type GetOrderStatusResponse = Order

// GetOrderStatus returns the current state of req.OrderID. This is the
// last step of every payment flow — poll it (e.g. every few seconds)
// until Status.IsTerminal() is true.
//
// Example:
//
//	for {
//		resp, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{OrderID: orderID})
//		if err != nil {
//			return err
//		}
//		if resp.Data.Status.IsTerminal() {
//			fmt.Println("final status:", resp.Data.StatusLabel)
//			break
//		}
//		time.Sleep(3 * time.Second)
//	}
func (c *Client) GetOrderStatus(ctx context.Context, req *GetOrderStatusRequest) (*PaymentResponse[GetOrderStatusResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/orders/%s", url.PathEscape(req.OrderID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[GetOrderStatusResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
