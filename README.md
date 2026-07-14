# gaian-go

Go client for the [Gaian Network](https://gaian.network) **v2 API** — a single signed gateway for user onboarding, KYC, wallets, and crypto-to-fiat payments across Vietnam, Thailand, and other supported markets.

## Installation

```bash
go get github.com/gaiannetwork/gaian-go
```

## Quick start

```go
import gaian "github.com/gaiannetwork/gaian-go"

client, err := gaian.NewClient("https://sandbox.gaian.network", apiKey, secretKey)
if err != nil {
    log.Fatal(err)
}

resp, err := client.CreateUser(ctx, &gaian.CreateUserRequest{})
if err != nil {
    log.Fatal(err)
}
userID := resp.Data.UserID
```

Get your API key **and secret** from the [Client Admin Portal](https://client-admin.gaian.network/). Both are shown only once at creation — store them securely, server-side only.

Every method hangs directly off `*Client` — there are no per-resource sub-clients (`client.CreateUser`, `client.Quote`, `client.PlaceOrder`, ...), not `client.User.Register(...)`.

---

## Environments

| Environment | Base URL |
|---|---|
| Sandbox | `https://sandbox.gaian.network` |
| Production | `https://api.gaian.network` |

---

## Authentication

Every request is HMAC-signed — there's no static API-key header. Each key has a **public key** (`apiKey`) and a **secret** (`secretKey`); `NewClient` takes both and signs every outgoing request automatically:

```go
message   = apiKey + method + path + query + canonicalBody + timestamp
signature = base64url( HMAC_SHA256(secretKey, message) )
```

where `canonicalBody` is the JSON body with top-level keys sorted alphabetically and compact-serialized (empty string for GET/no body). This is implemented in the `signer` subpackage — you never need to call it directly, `NewClient`/`*Client` handle signing for you.

```go
client, err := gaian.NewClient(
    "https://sandbox.gaian.network",
    os.Getenv("GAIAN_API_KEY"),
    os.Getenv("GAIAN_API_SECRET"),
)
```

> **Never** hardcode the API key/secret or expose them in client-side code. Signatures must be generated on your backend.

---

## Payment flows

### Standard flow (on-chain signing required)

```
Quote (or ParseQR first) → PlaceOrder → sign & broadcast the returned transaction
→ VerifyOrder → GetOrderStatus (poll until terminal)
```

### Prefunded flow (no on-chain signing, settles from tenant balance)

```
QuotePrefund → PlacePrefundedOrder → GetOrderStatus (poll)
```

### Direct bank-transfer flow (production only, no QR code)

```
ListBanks → VerifyAccount (recommended) → QuoteDirect (or QuoteDirectPrefund)
→ PlaceOrder (or PlacePrefundedOrder) → VerifyOrder (on-chain variant only) → GetOrderStatus
```

A `QuoteDirect`/`Quote` quote must be consumed by `PlaceOrder`; a `QuoteDirectPrefund`/`QuotePrefund` quote must be consumed by `PlacePrefundedOrder` — cross-use is rejected by the API. Quotes are single-use with a short TTL (~120s).

---

## Usage

### Users

#### Create a user

```go
resp, err := client.CreateUser(ctx, &gaian.CreateUserRequest{})
userID := resp.Data.UserID
```

No wallet or KYC data is attached at creation time — link a wallet and submit KYC afterward using the returned `userID`.

#### Get a user by ID or wallet

```go
resp, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: userID})

resp, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{WalletAddress: "0xWalletAddress"})
```

#### Check market access

```go
resp, err := client.GetMarkets(ctx, &gaian.GetMarketsRequest{UserID: userID})
if resp.Data.Markets["VN"].Status != gaian.MarketApproved {
    fmt.Println("Vietnam market not available")
}
```

| `MarketStatus` | Description |
|---|---|
| `gaian.MarketApproved` | Approved for payments in this market |
| `gaian.MarketPending` | KYC/verification in progress |
| `gaian.MarketNotStarted` | Nothing submitted yet for this market |
| `gaian.MarketRejected` | Rejected — check `RejectReason` (populated for PH/BR only) |

#### List / get orders

```go
resp, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{
    UserID:   userID,
    Page:     1,
    PageSize: 20,
    Status:   "completed", // optional filter
})
for _, o := range resp.Items {
    fmt.Println(o.OrderID, o.Status, o.FiatAmount, o.FiatCurrency)
}
fmt.Println("total:", resp.Pagination.TotalCount)

order, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{UserID: userID, OrderID: orderID})
```

---

### Wallets

```go
wallet, err := client.CreateWallet(ctx, &gaian.CreateWalletRequest{
    UserID:  userID,
    Address: "0xWalletAddress",
    Chain:   "base",
})
// the first wallet added for a user becomes primary automatically

resp, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{UserID: userID})
for _, w := range *resp.Data {
    fmt.Println(w.Address, w.Chain, w.IsPrimary)
}
```

---

### KYC

#### Hosted KYC link

The returned URL is time-limited. Generate it on demand — do not cache.

```go
resp, err := client.GenerateKYCLink(ctx, &gaian.GenerateKYCLinkRequest{
    UserID:      userID,
    CallbackURL: "https://myapp.com/kyc/callback", // optional
})
// Redirect the user to resp.Data.URL
```

#### Submit KYC information directly

```go
resp, err := client.SubmitKYC(ctx, &gaian.SubmitKYCRequest{
    UserID:             userID,
    FirstName:          "Nguyen",
    LastName:           "Van A",
    Email:              "user@example.com",
    Gender:             "male",
    DateOfBirth:        "1990-01-15",
    Nationality:        "VN",
    NationalID:         "012345678901",
    Type:               "national_id",
    ExpiryDate:         "2025-01-01",
    AddressLine1:       "123 Le Loi",
    City:               "Ho Chi Minh City",
    CountryOfResidence: "VN",
    Occupation:         "OCC9",
    PhoneCountryCode:   "+84",
    PhoneNumber:        "912345678",
    FrontIDImage:       "<base64>",
    BackIDImage:        "<base64>",
    HoldIDImage:        "<base64>",
})
fmt.Println(resp.Data.KYCStatus)
```

#### Update KYC information

All fields are optional pointers — only non-nil fields are sent/updated.

```go
newEmail := "updated@example.com"
resp, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{
    UserID: userID,
    Email:  &newEmail,
})
```

---

### Organization

```go
resp, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{})
fmt.Printf("%s %s available\n", resp.Data.AvailableBalance, resp.Data.Currency)
```

Restricted to tenants on the prefund billing model.

---

### Payments

#### Parse a QR code

```go
resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: qrString})
// Country is an optional ISO-3166-1 alpha-2 hint:
resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: qrString, Country: "VN"})

fmt.Println(resp.Data.BankBin, resp.Data.AccountNumber)
```

#### Create a quote

```go
resp, err := client.Quote(ctx, &gaian.QuoteRequest{
    QRString:           qrString,
    Amount:             500000, // fiat amount
    ChainID:            gaian.ChainSolana,
    SettlementCurrency: "USDC",
    UserID:             userID,
})
quoteID := resp.Data.QuoteID
```

| `ChainID` | Chain |
|---|---|
| `gaian.ChainEthereum` | Ethereum (1) |
| `gaian.ChainOptimism` | Optimism (10) |
| `gaian.ChainBSC` | BSC (56) |
| `gaian.ChainPolygon` | Polygon (137) |
| `gaian.ChainBase` | Base (8453) |
| `gaian.ChainArbitrum` | Arbitrum (42161) |
| `gaian.ChainSolana` | Solana (101) |

Prefunded quote (no on-chain step, no `ChainID`):

```go
resp, err := client.QuotePrefund(ctx, &gaian.QuotePrefundRequest{
    QRString:           qrString,
    Amount:             500000,
    SettlementCurrency: "USDC",
    UserID:             userID,
})
```

Direct bank-transfer quote (no QR code, production only — see `ListBanks`/`VerifyAccount` below):

```go
resp, err := client.QuoteDirect(ctx, &gaian.QuoteDirectRequest{
    AccountNumber:      "0123456789",
    Code:               bankCode, // from ListBanks
    Amount:             500000,
    Country:            "VN",
    ChainID:            gaian.ChainSolana,
    SettlementCurrency: "USDC",
    UserID:             userID, // exactly one of UserID/WalletAddress/Email
})

// or, prefunded:
resp, err := client.QuoteDirectPrefund(ctx, &gaian.QuoteDirectPrefundRequest{ /* same, minus ChainID */ })
```

#### Place an order

```go
resp, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{QuoteID: quoteID})

// resp.Data.EncodedTransaction is what you sign and broadcast on-chain.
txHash := broadcastTransaction(resp.Data.EncodedTransaction)
```

Prefunded order (no on-chain step — go straight to polling `GetOrderStatus`):

```go
resp, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{QuoteID: quoteID})
```

#### Verify the on-chain transaction

```go
resp, err := client.VerifyOrder(ctx, &gaian.VerifyOrderRequest{
    OrderID:         orderID,
    TransactionHash: txHash,
})
```

> **The API returns HTTP 200 even when verification did not succeed.** A `nil` error does **not** mean the payment is confirmed — always branch on `resp.Data.Status`:

```go
switch resp.Data.Status {
case gaian.VerifyVerified:
    // confirmed — poll GetOrderStatus for settlement
case gaian.VerifyPending:
    // transaction hasn't propagated on-chain yet — resubmit the same TransactionHash shortly
case gaian.VerifyFailed:
    log.Printf("verification failed: %s", resp.Data.Message)
}
```

#### Poll order status

```go
for {
    resp, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{OrderID: orderID})
    if err != nil {
        log.Fatal(err)
    }
    if resp.Data.Status.IsTerminal() {
        fmt.Println("final status:", resp.Data.StatusLabel)
        break
    }
    time.Sleep(3 * time.Second)
}
```

| `OrderStatus` | Label | Terminal? |
|---|---|---|
| `gaian.OrderPending` (0) | `pending` | no |
| `gaian.OrderAwaitingDeposit` (1) | `awaiting_deposit` | no |
| `gaian.OrderPaymentReceived` (2) | `payment_received` | no |
| `gaian.OrderQueued` (3) | `queued` | no |
| `gaian.OrderProcessing` (4) | `processing` | no |
| `gaian.OrderCompleted` (10) | `completed` | **yes** |
| `gaian.OrderFailed` (20) | `failed` | **yes** |
| `gaian.OrderCancelled` (21) | `cancelled` | **yes** |
| `gaian.OrderExpired` (22) | `expired` | **yes** |
| `gaian.OrderUnknown` (99) | `unknown` | no |

`VerifyStatus` (returned by `VerifyOrder`) is a **separate type** from `OrderStatus` — their integer values overlap (both have a `1`) but mean different things. Don't compare across the two.

---

### Banks (direct bank transfer)

Production only — not available in sandbox.

```go
resp, err := client.ListBanks(ctx, &gaian.ListBanksRequest{Country: "VN"})
for _, b := range resp.Data.Banks {
    fmt.Println(b.Code, b.Name) // b.Code is what QuoteDirect/VerifyAccount expect
}

verify, err := client.VerifyAccount(ctx, &gaian.VerifyAccountRequest{
    AccountNumber: "0123456789",
    Code:          bankCode,
    Country:       "VN",
})
```

> A failed verification (bad account number) is **not** an error — it returns HTTP 200 with `Valid: false` and `AccountName: nil`. Check `resp.Data.Valid`, not just the returned error.

---

## Response envelopes

Every method returns one of two generic wrappers (`response.go`):

- **`UserResposne[T]`** — `{data, requestId}` — used by User/KYC/Wallet/Organization endpoints.
- **`PaymentResponse[T]`** — `{success, requestId, data}` — used by Payment endpoints (`ParseQR`, `Quote*`, `PlaceOrder*`, `VerifyOrder`, `GetOrderStatus`, `ListBanks`, `VerifyAccount`).

`ListUserOrders` is the one exception — its response has `data`/`pagination`/`requestId` as siblings at the top level, so it returns `*ListUserOrdersResponse` directly instead of a wrapped envelope.

Check each method's doc comment (`go doc gaian.<Method>`) if you're unsure which shape applies.

---

## Error handling

All methods return a standard `error`. Non-2xx responses are currently wrapped in `ErrUnexpectedStatus` (check with `errors.Is`):

```go
resp, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: "does-not-exist"})
if err != nil {
    if errors.Is(err, gaian.ErrUnexpectedStatus) {
        // inspect err.Error() for the status code and raw response body
    }
    log.Fatal(err)
}
```

`gaian.APIError` and its `IsNotFound`/`IsConflict`/`IsUnauthorized`/`IsForbidden` helpers exist in `errors.go` but aren't wired into the client yet — see the note at the top of that file.

---

## Configuration

```go
client, err := gaian.NewClient(baseURL, apiKey, secretKey,
    gaian.WithHTTPClient(myHTTPClient), // any type implementing gaian.HTTPClient
    gaian.WithLogger(myLogger),         // optional: implement gaian.Logger (Info/Error/Debug)
    gaian.WithDebug(true),              // log request/response via the logger
)
```

`WithHTTPClient` is also how you inject timeouts — pass an `*http.Client` with `Timeout` set, or any custom `HTTPClient` implementation.

---

## License

MIT
