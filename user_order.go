package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GetUserOrderRequest identifies a single order belonging to a user.
type GetUserOrderRequest struct {
	UserID  string `param:"userId"`
	OrderID string `param:"orderId"`
}

// GetUserOrderResponse is the same shape as an item in
// ListUserOrdersResponse.Items.
type GetUserOrderResponse = UserOrder

// GetUserOrder returns a single order for req.UserID by req.OrderID. Use
// ListUserOrders if you need to browse/page through a user's order
// history instead.
//
// Example:
//
//	resp, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{
//		UserID:  userID,
//		OrderID: orderID,
//	})
func (c *Client) GetUserOrder(ctx context.Context, req *GetUserOrderRequest) (*UserResposne[GetUserOrderResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/orders/%s", url.PathEscape(req.UserID), url.PathEscape(req.OrderID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GetUserOrderResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
