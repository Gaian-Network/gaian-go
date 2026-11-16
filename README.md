# gaian-go

Go client for the [Gaian Network](https://gaian.network) **v2 API** — a single signed gateway for user onboarding, KYC, wallets, and crypto-to-fiat payments across Vietnam, Thailand, and other supported markets.

## Table of contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Environments](#environments)
- [Authentication](#authentication)
- [Payment flows](#payment-flows)
  - [Standard flow (on-chain signing required)](#standard-flow-on-chain-signing-required)
  - [Prefunded flow (no on-chain signing, settles from tenant balance)](#prefunded-flow-no-on-chain-signing-settles-from-tenant-balance)
  - [Direct bank-transfer flow (production only, no QR code)](#direct-bank-transfer-flow-production-only-no-qr-code)
- [Usage](#usage)
  - [Users](#users)
    - [Create a user](#create-a-user)
    - [Get a user by ID or wallet](#get-a-user-by-id-or-wallet)
    - [Check market access](#check-market-access)
    - [KYC requirements per market](#kyc-requirements-per-market)
    - [Payment methods](#payment-methods)
    - [List / get orders](#list--get-orders)
  - [Wallets](#wallets)
  - [KYC](#kyc)
    - [Hosted KYC link](#hosted-kyc-link)
    - [Submit KYC information directly](#submit-kyc-information-directly)
    - [Update KYC information](#update-kyc-information)
    - [Import a shared KYC session](#import-a-shared-kyc-session)
  - [Organization](#organization)
  - [Payments](#payments)
    - [Parse a QR code](#parse-a-qr-code)
    - [Create a quote](#create-a-quote)
    - [Place an order](#place-an-order)
    - [Verify the on-chain transaction](#verify-the-on-chain-transaction)
    - [Poll order status](#poll-order-status)
    - [Exchange rates](#exchange-rates)
    - [User limits](#user-limits)
  - [Banks (direct bank transfer)](#banks-direct-bank-transfer)
- [Webhooks](#webhooks)
- [Response envelopes](#response-envelopes)
- [Error handling](#error-handling)
- [Configuration](#configuration)
- [License](#license)

---

## Installation

```bash
go get github.com/Gaian-Network/gaian-go/v2
```

## Quick start

```go
import gaian "github.com/Gaian-Network/gaian-go/v2"

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

#### KYC requirements per market

```go
resp, err := client.GetMarketRequirements(ctx, &gaian.GetMarketRequirementsRequest{UserID: userID})
vn := resp.Data.Markets["VN"]
fmt.Println(vn.RequiredFields, vn.MissingFields)
```

Field names are the JSON names of `SubmitKYCRequest` (for example `nationalId`, `kycReport`). `MissingFields` is what the user still has to provide for that market.

#### Payment methods

```go
resp, err := client.GetPaymentMethods(ctx, &gaian.GetPaymentMethodsRequest{UserID: userID})
for _, m := range *resp.Data {
    fmt.Println(m.Method, m.Name, m.Status) // gaian.PaymentMethodQR / gaian.PaymentMethodCard
}
```

`Status` uses the same values as `MarketStatus`. `RejectReason` and `Action` are `nil` unless the method is rejected.

#### List / get orders

```go
resp, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{
    UserID:   userID,
    Page:     1,
    PageSize: 20, // default 50, at most 100
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
frontImage, backImage, holdImage := "<base64>", "<base64>", "<base64>"

resp, err := client.SubmitKYC(ctx, &gaian.SubmitKYCRequest{
    UserID:             userID,
    FirstName:          "Nguyen",
    LastName:           "Van A",
    Email:              "user@example.com",
    Gender:             "male",
    DateOfBirth:        "1990-01-15",
    Nationality:        "VN",
    NationalID:         "012345678901",
    Type:               gaian.DocumentTypeIDCard,
    ExpiryDate:         "2025-01-01",
    AddressLine1:       "123 Le Loi",
    City:               "Ho Chi Minh City",
    CountryOfResidence: "VN",
    Occupation:         "OCC9",
    PhoneCountryCode:   "+84",
    PhoneNumber:        "912345678",
    FrontIDImage:       &frontImage,
    BackIDImage:        &backImage,
    HoldIDImage:        &holdImage,
})
fmt.Println(resp.Data.KYCStatus)
```

`FrontIDImage`/`BackIDImage`/`HoldIDImage`/`KYCReport` are `*string` — leave them `nil` to omit entirely (some flows don't require ID images at submission time). `KYCReport` is required by some markets; `GetMarketRequirements` lists `kycReport` when it is.

`Gender`, `PhoneCountryCode` and `PhoneNumber` are optional: an empty value is not sent.

| `DocumentType` | Value |
|---|---|
| `gaian.DocumentTypeIDCard` | `ID_CARD` |
| `gaian.DocumentTypePassport` | `PASSPORT` |
| `gaian.DocumentTypeDriversLicense` | `DRIVERS_LICENSE` |
| `gaian.DocumentTypeResidencePermit` | `RESIDENCE_PERMIT` |

#### Update KYC information

All fields are optional pointers — only non-nil fields are sent/updated.

```go
newEmail := "updated@example.com"
resp, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{
    UserID: userID,
    Email:  &newEmail,
})
```

#### Import a shared KYC session

For a user who already passed KYC with a partner application (reusable KYC). The partner gives you a share token:

```go
resp, err := client.ImportSharedKYC(ctx, &gaian.ImportSharedKYCRequest{
    UserID:     userID,
    ShareToken: shareToken,
})
fmt.Println(resp.Data.SessionID, resp.Data.Status)
```

Returns 400 for an invalid or expired token, 404 for an unknown user, 409 when the user has already submitted KYC.

---

### Organization

```go
resp, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{})
fmt.Printf("%s %s available\n", resp.Data.AvailableBalance, resp.Data.Currency)
```

Restricted to tenants on the prefund billing model. `resp.Data.WalletAddress` and `resp.Data.Chain` are `*string`: the prefund deposit wallet, or `nil` when none is configured.

---

### Payments

#### Parse a QR code

```go
resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: qrString})
// Country is an optional ISO-3166-1 alpha-2 hint:
resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: qrString, Country: "VN"})

fmt.Println(resp.Data.BankBin, resp.Data.AccountNumber)
```

QR codes from Brazil, Argentina, Peru and Bolivia need the payer's identity (one of `UserID`, `WalletAddress`, `Email`); without it the API answers 400 with code `IDENTITY_REQUIRED`. Peru also takes `Network` (`"YAPE"` or `"PLIN"`):

```go
resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{
    QRString: qrString,
    Country:  "PE",
    Network:  "YAPE",
    UserID:   userID,
})
```

`resp.Data.BankName` is a `*string`, `nil` when the QR carries no bank name.

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

Identify the payer with exactly one of `UserID`, `WalletAddress` or `Email`. The SDK does not check this; the API answers 400 when none or more than one is set. Optional fields: `Network` (`"YAPE"` or `"PLIN"`, Peru only) and `FromAddress` (the on-chain wallet the deposit must come from, when it is not a wallet linked to the user).

In the response, `SettlementAmount` = `SettlementBaseUSD` + `ProtocolFeeUSD` + `MinimumFeeUSD` + `TenantFeeUSD` (all decimal strings).

| `ChainID` | Chain |
|---|---|
| `gaian.ChainEthereum` | Ethereum (1) |
| `gaian.ChainOptimism` | Optimism (10) |
| `gaian.ChainBSC` | BSC (56) |
| `gaian.ChainPolygon` | Polygon (137) |
| `gaian.ChainBase` | Base (8453) |
| `gaian.ChainArbitrum` | Arbitrum (42161) |
| `gaian.ChainSolana` | Solana (101) |
| `gaian.ChainStellar` | Stellar (148) |
| `gaian.ChainAvalanche` | Avalanche (43114) |
| `gaian.ChainArc` | Arc (5042002) |

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

The response also carries `DepositAddress`, `TokenAddress`, `SettlementAmount`, `SettlementAmountBaseUnits`, `SettlementCurrency` and `Memo`. Attach `Memo` to the deposit when it is not empty (chains such as Stellar need it).

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

#### Exchange rates

```go
resp, err := client.GetExchangeRates(ctx, &gaian.GetExchangeRatesRequest{
    Country:            "VN",   // optional: empty returns every supported country
    SettlementCurrency: "USDC", // optional: the API defaults to USDC
})
for _, r := range resp.Data.Rates {
    fmt.Println(r.Country, r.FiatCurrency, r.ExchangeRate)
}
```

These rates are indicative (`IndicativeOnly` is always `true`). Only a quote binds a rate.

#### User limits

```go
resp, err := client.GetUserLimits(ctx, &gaian.GetUserLimitsRequest{
    UserID:  userID,
    Country: "VN", // optional: empty returns every configured country
})
for _, c := range resp.Data.Countries {
    fmt.Println(c.Country, c.CanTransact, c.MaxNextTransactionUSD)
}
```

`MaxNextTransactionUSD` is the largest order the user can place right now. `MaxPerDayUSD`, `DailyRemainingUSD` and `Weekly` are `nil` when that limit does not apply. USD figures here are `float64` because the API sends JSON numbers: use them for display and pre-checks, not for accounting.

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

## Webhooks

Gaian signs every webhook delivery with the endpoint's secret. Verify it before trusting the payload, and pass the body exactly as received (do not decode and re-encode it):

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "cannot read body", http.StatusBadRequest)
        return
    }

    err = gaian.VerifyWebhook(gaian.VerifyWebhookParams{
        Secret:    webhookSecret,
        Signature: r.Header.Get("X-Gaian-Signature"),
        Timestamp: r.Header.Get("X-Gaian-Timestamp"),
        Body:      body,
    })
    if err != nil {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    // body is authentic: decode it and handle the event
    w.WriteHeader(http.StatusOK)
}
```

`VerifyWebhook` returns `gaian.ErrInvalidWebhookSignature` when the signature does not match, and `gaian.ErrWebhookTimestampOutOfTolerance` when it matches but the timestamp is more than 300 seconds from now, in the past or the future. Change the window with `ToleranceSeconds`. The comparison is constant-time.

---

## Response envelopes

Every method returns one of two generic wrappers (`response.go`):

- **`UserResponse[T]`** — `{data, requestId}` — used by User/KYC/Wallet/Organization endpoints (including `ImportSharedKYC`, `GetMarketRequirements`, `GetPaymentMethods`). `UserResposne[T]` is a deprecated alias of it.
- **`PaymentResponse[T]`** — `{success, requestId, data}` — used by Payment endpoints (`ParseQR`, `Quote*`, `PlaceOrder*`, `VerifyOrder`, `GetOrderStatus`, `ListBanks`, `VerifyAccount`, `GetExchangeRates`, `GetUserLimits`).

`ListUserOrders` is the one exception — its body is the `{data, requestId}` envelope with a `pagination` sibling, so it returns `*ListUserOrdersResponse` directly (`Items` is the `data` array).

Check each method's doc comment (`go doc gaian.<Method>`) if you're unsure which shape applies.

---

## Error handling

All methods return a standard `error`. Every response with a status of 400 or above is a `*gaian.APIError` carrying `StatusCode`, `Code` (Payment endpoints only), `Message`, `RequestID` and the raw `Body`:

```go
resp, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: "does-not-exist"})
if err != nil {
    var apiErr *gaian.APIError
    if errors.As(err, &apiErr) {
        log.Printf("status=%d code=%s request_id=%s", apiErr.StatusCode, apiErr.Code, apiErr.RequestID)
    }
    if gaian.IsNotFound(err) {
        // the user does not exist
    }
    log.Fatal(err)
}
```

`IsNotFound`, `IsConflict`, `IsUnauthorized` and `IsForbidden` also work on wrapped errors. `errors.Is(err, gaian.ErrUnexpectedStatus)` stays true for every `*APIError`.

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
