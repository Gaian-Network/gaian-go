# gaian-go

Go client for the [Gaian Network](https://gaian.network) API — payment processing across Vietnam, Philippines, and Brazil via QR code and stablecoin settlement.

## Installation

```bash
go get github.com/gaiannetwork/gaian-go
```

Requires Go 1.21+. No external dependencies.

## Quick start

```go
import gaian "github.com/gaiannetwork/gaian-go"

c := gaian.New("your-api-key", gaian.Sandbox)
```

Get your API key from the [Client Admin Portal](https://client-admin.gaian-dev.network/). Keys are shown only once at creation — store them securely.

---

## Environments

| Constant | User API | Payment API |
|----------|----------|-------------|
| `gaian.Sandbox` | `dev-user.gaian-dev.network` | `dev-payments.gaian-dev.network` |
| `gaian.Production` | `user.gaian-dev.network` | `payments.gaian-dev.network` |

---

## Authentication

All requests require the `x-api-key` header. The client sets this automatically from the key passed to `New`.

```go
// Sandbox
c := gaian.New(os.Getenv("GAIAN_API_KEY"), gaian.Sandbox)

// Production
c := gaian.New(os.Getenv("GAIAN_API_KEY"), gaian.Production)
```

> **Never** hardcode API keys or expose them in client-side code.

---

## Payment flows

### Standard flow (on-chain signing required)

```
Register → KYC → Check Policy → Parse QR → Calculate Exchange
→ Place Order → Sign & Broadcast Transaction → Verify Order → Poll Status
```

### Prefunded flow (no on-chain signing)

```
Place Prefunded Order → Poll Status
```

---

## Usage

### User

#### Register a user

The endpoint is idempotent — returns the existing user if the wallet is already registered.

```go
user, err := c.User.Register(ctx, "0xWalletAddress")
if err != nil && !gaian.IsConflict(err) {
    log.Fatal(err)
}
fmt.Println(user.ID, user.KYCStatus)
```

#### Get user by ID or wallet

```go
user, err := c.User.GetByID(ctx, 42)

user, err := c.User.GetByWallet(ctx, "0xWalletAddress")
```

#### Check market availability

```go
// Returns map[countryCode]MarketInfo, e.g. map["VN"]
markets, err := c.User.GetMarkets(ctx, "0xWalletAddress")
if markets["VN"].Status != gaian.MarketApproved {
    fmt.Println("Vietnam market not available:", markets["VN"].RejectReason)
}
```

#### List orders

```go
list, err := c.User.ListOrdersByWallet(ctx, "0xWalletAddress", gaian.OrderListOptions{
    Page:   1,
    Limit:  20,
    Status: gaian.OrderCompleted, // optional filter
})

for _, order := range list.Items {
    fmt.Println(order.OrderID, order.Status, order.FiatAmount, order.FiatCurrency)
}
fmt.Println("total:", list.Pagination.Total)
```

---

### KYC

#### Generate a Sumsub KYC link

The returned URL is time-limited. Generate it on demand — do not cache.

```go
url, err := c.User.GenerateKYCLink(ctx, "0xWalletAddress")
// Redirect user to url
```

#### Submit KYC information directly

```go
status, err := c.User.SubmitKYC(ctx, gaian.SubmitKYCRequest{
    UserWalletAddress: "0xWalletAddress",
    UserEmail:         "user@example.com",
    FirstName:         "Nguyen",
    LastName:          "Van A",
    DateOfBirth:       "1990-01-15",
    Gender:            "male",
    Nationality:       "VN",
    Type:              "national_id",
    NationalID:        "012345678901",
    IssueDate:         "2015-01-01",
    ExpiryDate:        "2025-01-01",
    AddressLine1:      "123 Le Loi",
    City:              "Ho Chi Minh City",
    State:             "HCM",
    ZipCode:           "70000",
    FrontIDImage:      "<base64>",
    BackIDImage:       "<base64>",
    HoldIDImage:       "<base64>",
    PhoneNumber:       "912345678",
    PhoneCountryCode:  "+84",
})
```

---

### Policy

Check whether a wallet can make a payment in a given market and get transaction limits.

```go
policy, err := c.Policy.CheckPolicy(ctx, "0xWalletAddress", "VN")
if !policy.IsAllowed {
    log.Fatalf("payment blocked: %v", policy.Reason)
}

fmt.Printf("tier: %s  per-tx: $%.0f  per-day: $%.0f\n",
    policy.Tier, policy.Limits.PerTransaction, policy.Limits.PerDay)
```

| Tier | Description |
|------|-------------|
| `gaian.TierKYC` | KYC-verified, higher limits |
| `gaian.TierNonKYC` | Unverified, lower limits |

---

### Payment

#### Parse a QR code

Supports EMVCO QR codes from Vietnam, Philippines, Thailand, and Brazil.

```go
info, err := c.Payment.ParseQR(ctx, qrString)
// Optional country hint:
info, err := c.Payment.ParseQR(ctx, qrString, gaian.ParseQROptions{Country: "VN"})

// Amount, Currency, BeneficiaryName are nullable — check before use
if info.Amount != nil && info.Currency != nil {
    fmt.Printf("%.0f %s\n", *info.Amount, *info.Currency)
}
if info.BeneficiaryName != nil {
    fmt.Println(*info.BeneficiaryName)
}
```

#### Calculate exchange rate

```go
exchange, err := c.Payment.CalculateExchange(ctx, gaian.CalculateExchangeRequest{
    Amount:  500000, // fiat amount
    Country: "VN",
    Chain:   gaian.Solana,
    Token:   gaian.USDC,
})

fmt.Printf("need %s %s (rate: %s, fee: %s)\n",
    exchange.CryptoAmount, exchange.CryptoCurrency,
    exchange.ExchangeRate, exchange.FeeAmount)
```

#### Place a standard order

```go
order, err := c.Payment.PlaceOrder(ctx, gaian.PlaceOrderRequest{
    QRString:       qrString,
    Amount:         500000,
    CryptoCurrency: gaian.USDC,
    FromAddress:    "0xWalletAddress",
    FiatCurrency:   gaian.VND,
    Chain:          gaian.Solana,
})

// order.CryptoTransferInfo contains destination wallet and amount.
// Build and sign the Solana transaction (legacy format, NOT VersionedTransaction),
// then call VerifyOrder with the transaction signature.
```

#### Verify the on-chain transaction

```go
result, err := c.Payment.VerifyOrder(ctx, order.OrderID, txSignature)
fmt.Println(result.Status, result.BankTransferStatus)
```

#### Place a prefunded order

No on-chain signing required. Uses the tenant's prefunded balance.

```go
order, err := c.Payment.PlacePrefundedOrder(ctx, gaian.PlacePrefundedOrderRequest{
    QRString:       qrString,
    Amount:         500000,
    CryptoCurrency: gaian.USDC,
    FromAddress:    "0xWalletAddress",
    FiatCurrency:   gaian.VND,
})
```

#### Poll order status

Poll until the order reaches a terminal state (`completed` or `failed`).

```go
for {
    status, err := c.Payment.GetOrderStatus(ctx, order.OrderID)
    if err != nil {
        log.Fatal(err)
    }

    if status.Status == gaian.OrderCompleted || status.Status == gaian.OrderFailed {
        fmt.Println("final status:", status.Status)
        break
    }
    time.Sleep(3 * time.Second)
}
```

| Status | Description |
|--------|-------------|
| `OrderAwaitingCryptoTransfer` | Waiting for on-chain transaction |
| `OrderVerified` | Transaction confirmed, bank transfer queued |
| `OrderProcessing` | Bank transfer in progress |
| `OrderCompleted` | Payment successful |
| `OrderFailed` | Payment failed |

---

### Tenant

```go
// Prefunded balance
balance, err := c.Tenant.GetBalance(ctx, gaian.USDC)
fmt.Printf("%.4f %s available\n", balance.AvailableBalance, balance.Currency)

// Spend cap usage
spend, err := c.Tenant.GetTotalSpent(ctx)
fmt.Printf("spent $%.2f of $%.2f limit\n", spend.TotalSpent, spend.Limit)
```

---

## Error handling

All methods return a standard `error`. API errors are typed as `*gaian.APIError` and expose `StatusCode`.

```go
user, err := c.User.GetByID(ctx, 999)
if err != nil {
    switch {
    case gaian.IsNotFound(err):
        // 404
    case gaian.IsConflict(err):
        // 409 — resource already exists
    case gaian.IsUnauthorized(err):
        // 401 — invalid or missing API key
    case gaian.IsForbidden(err):
        // 403 — insufficient permissions
    default:
        log.Fatal(err)
    }
}
```

---

## Configuration

```go
c := gaian.New(apiKey, gaian.Sandbox,
    gaian.WithTimeout(15*time.Second),
    gaian.WithHTTPClient(myHTTPClient), // any type implementing HTTPClient interface
    gaian.WithLogger(myLogger),         // optional: implement gaian.Logger
    gaian.WithDebug(true),              // log request/response to logger
)
```

---

## Supported currencies and chains

| Type | Values |
|------|--------|
| Fiat | `VND` (Vietnam), `PHP` (Philippines), `BRL` (Brazil) |
| Crypto | `USDC`, `USDT` |
| Chains | `Solana`, `Ethereum`, `Polygon`, `Arbitrum`, `Base` |

> Currently only Solana is fully supported for exchange calculation and prefunded orders.

---

## License

MIT
