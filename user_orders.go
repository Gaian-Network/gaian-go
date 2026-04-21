package gaian

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// OrderListOptions controls pagination and optional status filtering.
type OrderListOptions struct {
	Page   int         // min 1, default 1
	Limit  int         // min 1, max 100, default 20
	Status OrderStatus // leave empty to return all statuses
}

func (o OrderListOptions) toQuery() url.Values {
	q := url.Values{}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Status != "" {
		q.Set("status", string(o.Status))
	}
	return q
}

type ordersData struct {
	User   User      `json:"user"`
	Orders OrderList `json:"orders"`
}

// ListOrdersByWallet returns paginated orders for the given wallet address.
func (s *UserService) ListOrdersByWallet(ctx context.Context, walletAddress string, opts OrderListOptions) (*OrderList, error) {
	var resp envelope[ordersData]
	path := fmt.Sprintf("/api/v1/users/wallet/%s/orders", url.PathEscape(walletAddress))
	if err := s.c.get(ctx, s.c.userBaseURL, path, opts.toQuery(), &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &resp.Data.Orders, nil
}

// ListOrdersByIdentifier returns paginated orders for a user identified by
// their numeric ID or email address.
func (s *UserService) ListOrdersByIdentifier(ctx context.Context, identifier string, opts OrderListOptions) (*OrderList, error) {
	var resp envelope[ordersData]
	path := fmt.Sprintf("/api/v1/users/%s/orders", url.PathEscape(identifier))
	if err := s.c.get(ctx, s.c.userBaseURL, path, opts.toQuery(), &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &resp.Data.Orders, nil
}
