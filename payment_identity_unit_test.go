package gaian_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

const quoteOKBody = `{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","route":{"routeId":1},"fiatAmount":500000,` +
	`"fiatCurrency":"VND","settlementAmount":"21.0","settlementBaseUsd":"20.4","settlementCurrency":"USDC","chainId":148,` +
	`"exchangeRate":"24390","protocolFeeUsd":"0.1","minimumFeeUsd":"0.05","tenantFeeUsd":"0.45","qrInfo":{"isValid":true,` +
	`"encodedString":"raw-qr","country":"VN","qrProvider":"napas","bankBin":"970436","accountNumber":"0123456789",` +
	`"beneficiaryName":"NGUYEN VAN A"},"expiresAt":"2030-01-01T00:00:00Z"}}`

func sentBody(t *testing.T, mock *mockHTTPClient) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	return body
}

// The SDK must not reject a request without an identity: the server owns
// that check and answers 400, so the missing-identity case only asserts
// that the call still goes out.
func TestQuote_IdentityAndSourceFields(t *testing.T) {
	t.Parallel()

	base := func() gaian.QuoteRequest {
		return gaian.QuoteRequest{
			QRString:           "raw-qr",
			Amount:             500000,
			ChainID:            gaian.ChainSolana,
			SettlementCurrency: "USDC",
		}
	}

	tests := []struct {
		name string
		mut  func(*gaian.QuoteRequest)
		want map[string]any
	}{
		{
			name: "wallet address",
			mut:  func(r *gaian.QuoteRequest) { r.WalletAddress = "0xabc" },
			want: map[string]any{"walletAddress": "0xabc"},
		},
		{
			name: "email with network and fromAddress",
			mut: func(r *gaian.QuoteRequest) {
				r.Email = "a@example.com"
				r.Network = "YAPE"
				r.FromAddress = "0xfrom"
			},
			want: map[string]any{"email": "a@example.com", "network": "YAPE", "fromAddress": "0xfrom"},
		},
		{
			name: "user id keeps working",
			mut:  func(r *gaian.QuoteRequest) { r.UserID = "usr_1" },
			want: map[string]any{"userId": "usr_1"},
		},
		{
			name: "no identity is sent as is",
			mut:  func(r *gaian.QuoteRequest) {},
			want: map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{response: mockResponse(http.StatusOK, quoteOKBody)} //nolint:bodyclose // response body closed by client
			client := newMockClient(t, mock)

			req := base()
			tc.mut(&req)
			if _, err := client.Quote(context.Background(), &req); err != nil {
				t.Fatalf("Quote: %v", err)
			}

			got := sentBody(t, mock)
			want := map[string]any{
				"qrString": "raw-qr", "amount": float64(500000), "chainId": float64(101), "settlementCurrency": "USDC",
			}
			for k, v := range tc.want {
				want[k] = v
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body = %v\nwant %v", got, want)
			}
		})
	}
}

func TestQuote_ParsesNewResponseFields(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{response: mockResponse(http.StatusOK, quoteOKBody)} //nolint:bodyclose // response body closed by client
	client := newMockClient(t, mock)

	resp, err := client.Quote(context.Background(), &gaian.QuoteRequest{QRString: "raw-qr", Amount: 1, WalletAddress: "0xabc"})
	if err != nil {
		t.Fatalf("Quote: %v", err)
	}
	if resp.Data.SettlementBaseUSD != "20.4" || resp.Data.TenantFeeUSD != "0.45" {
		t.Errorf("SettlementBaseUSD=%q TenantFeeUSD=%q", resp.Data.SettlementBaseUSD, resp.Data.TenantFeeUSD)
	}
	if resp.Data.ChainID != gaian.ChainStellar {
		t.Errorf("ChainID = %v, want Stellar", resp.Data.ChainID)
	}
}

func TestQuotePrefund_IdentityAndNetwork(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{response: mockResponse(http.StatusOK, quoteOKBody)} //nolint:bodyclose // response body closed by client
	client := newMockClient(t, mock)

	resp, err := client.QuotePrefund(context.Background(), &gaian.QuotePrefundRequest{
		QRString: "raw-qr", Amount: 1000, SettlementCurrency: "USDC", Email: "a@example.com", Network: "PLIN",
	})
	if err != nil {
		t.Fatalf("QuotePrefund: %v", err)
	}

	want := map[string]any{
		"qrString": "raw-qr", "amount": float64(1000), "settlementCurrency": "USDC",
		"email": "a@example.com", "network": "PLIN",
	}
	if got := sentBody(t, mock); !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v\nwant %v", got, want)
	}
	if resp.Data.TenantFeeUSD != "0.45" {
		t.Errorf("TenantFeeUSD = %q", resp.Data.TenantFeeUSD)
	}
}

func TestQuoteDirect_NetworkAndFromAddress(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","settlementBaseUsd":"9.9","tenantFeeUsd":"0.1"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.QuoteDirect(context.Background(), &gaian.QuoteDirectRequest{
		AccountNumber: "123", Code: "BCP", Amount: 100, Country: "PE", ChainID: gaian.ChainBase,
		SettlementCurrency: "USDC", Email: "a@example.com", Network: "YAPE", FromAddress: "0xfrom",
	})
	if err != nil {
		t.Fatalf("QuoteDirect: %v", err)
	}

	body := sentBody(t, mock)
	if body["network"] != "YAPE" || body["fromAddress"] != "0xfrom" || body["email"] != "a@example.com" {
		t.Errorf("body = %v", body)
	}
	if _, ok := body["userId"]; ok {
		t.Error("empty userId leaked into body")
	}
	if resp.Data.SettlementBaseUSD != "9.9" || resp.Data.TenantFeeUSD != "0.1" {
		t.Errorf("response fee fields = %+v", resp.Data)
	}
}

func TestQuoteDirectPrefund_NetworkSent(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1"}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.QuoteDirectPrefund(context.Background(), &gaian.QuoteDirectPrefundRequest{
		AccountNumber: "123", Code: "BCP", Amount: 100, Country: "PE", SettlementCurrency: "USDC",
		WalletAddress: "0xabc", Network: "PLIN",
	}); err != nil {
		t.Fatalf("QuoteDirectPrefund: %v", err)
	}

	body := sentBody(t, mock)
	if body["network"] != "PLIN" || body["walletAddress"] != "0xabc" {
		t.Errorf("body = %v", body)
	}
}

func TestParseQR_IdentityAndNetworkSent(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"isValid":true,"encodedString":"raw-qr","country":"PE",`+
				`"qrProvider":"yape","bankBin":"","accountNumber":"999","beneficiaryName":"JUAN","bankName":"BCP"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{
		QRString: "raw-qr", Country: "PE", Network: "YAPE", WalletAddress: "0xabc",
	})
	if err != nil {
		t.Fatalf("ParseQR: %v", err)
	}

	want := map[string]any{"qrString": "raw-qr", "country": "PE", "network": "YAPE", "walletAddress": "0xabc"}
	if got := sentBody(t, mock); !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v\nwant %v", got, want)
	}
	if resp.Data.BankName == nil || *resp.Data.BankName != "BCP" {
		t.Errorf("BankName = %v, want BCP", resp.Data.BankName)
	}
}

func TestParseQR_WithoutIdentityStillSentAndServer400Surfaces(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusBadRequest, //nolint:bodyclose // response body closed by client
			`{"success":false,"requestId":"req_1","error":{"code":"IDENTITY_REQUIRED","message":"identity required"}}`),
	}
	client := newMockClient(t, mock)

	_, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr", Country: "BR"})
	if err == nil {
		t.Fatal("expected the server's 400 to be returned")
	}
	if mock.lastReq == nil {
		t.Fatal("request never left the SDK")
	}
	if got := sentBody(t, mock); !reflect.DeepEqual(got, map[string]any{"qrString": "raw-qr", "country": "BR"}) {
		t.Errorf("body = %v", got)
	}
}

func TestPlaceOrder_ParsesSettlementFields(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":1,"statusLabel":"PENDING","chainId":43114,`+
				`"depositAddress":"0xdep","tokenAddress":"0xtok","settlementAmount":"21.0","settlementAmountBaseUnits":"21000000",`+
				`"settlementCurrency":"USDC","encodedTransaction":"enc","expiresAt":"2030-01-01T00:00:00Z","memo":"m-1"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.PlaceOrder(context.Background(), &gaian.PlaceOrderRequest{QuoteID: "quote_1"})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}

	d := resp.Data
	if d.TokenAddress != "0xtok" || d.SettlementAmount != "21.0" || d.SettlementAmountBaseUnits != "21000000" ||
		d.SettlementCurrency != "USDC" || d.Memo != "m-1" {
		t.Errorf("data = %+v", d)
	}
	if d.ChainID != gaian.ChainAvalanche {
		t.Errorf("ChainID = %v, want Avalanche", d.ChainID)
	}
}

func TestChainIDValuesMatchService(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		got  gaian.ChainID
		want int
	}{
		"Stellar":   {gaian.ChainStellar, 148},
		"Avalanche": {gaian.ChainAvalanche, 43114},
		"Arc":       {gaian.ChainArc, 5042002},
	} {
		if int(tc.got) != tc.want {
			t.Errorf("%s = %d, want %d", name, tc.got, tc.want)
		}
	}
}

func TestGetOrganizationBalance_ParsesWalletAndChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		json       string
		wantWallet *string
		wantChain  *string
	}{
		{
			name:       "set",
			json:       `"walletAddress":"0xabc","chain":"BSC"`,
			wantWallet: ptr("0xabc"),
			wantChain:  ptr("BSC"),
		},
		{name: "null", json: `"walletAddress":null,"chain":null`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
					`{"data":{"currency":"USDC","availableBalance":"1","totalSpent":"2","pendingSettlement":"3",`+tc.json+`},"requestId":"r"}`),
			}
			client := newMockClient(t, mock)

			resp, err := client.GetOrganizationBalance(context.Background(), &gaian.GetOrganizationBalanceRequest{})
			if err != nil {
				t.Fatalf("GetOrganizationBalance: %v", err)
			}
			if !reflect.DeepEqual(resp.Data.WalletAddress, tc.wantWallet) || !reflect.DeepEqual(resp.Data.Chain, tc.wantChain) {
				t.Errorf("WalletAddress=%v Chain=%v", resp.Data.WalletAddress, resp.Data.Chain)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
