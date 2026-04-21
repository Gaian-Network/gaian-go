package gaian

import "time"

// ── Enums ────────────────────────────────────────────────────────────────────

type KYCStatus string

const (
	KYCApproved   KYCStatus = "approved"
	KYCPending    KYCStatus = "pending"
	KYCNotStarted KYCStatus = "not_started"
	KYCRejected   KYCStatus = "rejected"
)

type OrderStatus string

const (
	OrderAwaitingCryptoTransfer OrderStatus = "awaiting_crypto_transfer"
	OrderVerified               OrderStatus = "verified"
	OrderProcessing             OrderStatus = "processing"
	OrderCompleted              OrderStatus = "completed"
	OrderFailed                 OrderStatus = "failed"
)

type MarketStatus string

const (
	MarketApproved   MarketStatus = "approved"
	MarketPending    MarketStatus = "pending"
	MarketNotStarted MarketStatus = "not_started"
	MarketRejected   MarketStatus = "rejected"
)

type FiatCurrency string

const (
	VND FiatCurrency = "VND"
	PHP FiatCurrency = "PHP"
	BRL FiatCurrency = "BRL"
)

type CryptoCurrency string

const (
	USDC CryptoCurrency = "USDC"
	USDT CryptoCurrency = "USDT"
)

type Chain string

const (
	Solana   Chain = "Solana"
	Ethereum Chain = "Ethereum"
	Polygon  Chain = "Polygon"
	Arbitrum Chain = "Arbitrum"
	Base     Chain = "Base"
)

type PolicyTier string

const (
	TierKYC    PolicyTier = "KYC"
	TierNonKYC PolicyTier = "NON_KYC"
)

// Route identifies the payment route provider.
type Route int

const (
	RouteBIDV    Route = 1
	RouteAlixPay Route = 2
)

// ── Core entities ─────────────────────────────────────────────────────────────

// User represents a Gaian platform user.
type User struct {
	ID            int        `json:"id"`
	Email         *string    `json:"email"`
	WalletAddress *string    `json:"walletAddress"`
	TenantID      *int       `json:"tenantId"`
	KYCStatus     KYCStatus  `json:"kycStatus"`
	FirstName     *string    `json:"firstName"`
	LastName      *string    `json:"lastName"`
	Nationality   *string    `json:"nationality"`
	DateOfBirth   *string    `json:"dateOfBirth"`
	PhoneNumber   *string    `json:"phoneNumber"`
	CreatedAt     *time.Time `json:"createdAt"` // nullable per API spec
	UpdatedAt     *time.Time `json:"updatedAt"` // nullable per API spec
}

// BankInfo holds bank identification inside a QR payload.
type BankInfo struct {
	BankBin    *string `json:"bankBin"`
	BankNumber *string `json:"bankNumber"`
}

// ProviderInfo identifies the payment provider inside a QR payload.
type ProviderInfo struct {
	GUID    *string `json:"guid"`
	Name    *string `json:"name"`
	Service *string `json:"service"`
}

// QRInfo is the parsed content of a QR payment code.
type QRInfo struct {
	BankInfo        BankInfo     `json:"bankInfo"`
	ProviderInfo    ProviderInfo `json:"providerInfo"`
	BeneficiaryName *string      `json:"beneficiaryName"`
	EncodedString   *string      `json:"encodedString"`
	AdditionalData  *string      `json:"additionalData"`
	CountryCode     *string      `json:"countryCode"`
	Amount          *float64     `json:"amount"`
	Purpose         *string      `json:"purpose"`
}

// ParsedQRInfo is the response shape from the parseQr endpoint.
type ParsedQRInfo struct {
	IsValid         bool           `json:"isValid"`
	EncodedString   string         `json:"encodedString"`
	Country         string         `json:"country"`
	QRProvider      *string        `json:"qrProvider"`
	BankBin         *string        `json:"bankBin"`
	AccountNumber   *string        `json:"accountNumber"`
	Amount          *float64       `json:"amount"`
	Currency        *string        `json:"currency"`
	Purpose         *string        `json:"purpose"`
	Nation          *string        `json:"nation"`
	BeneficiaryName *string        `json:"beneficiaryName"`
	DetailedQRInfo  map[string]any `json:"detailedQrInfo"`
}

// BankTransactionReference holds identifiers from the bank transfer leg.
type BankTransactionReference struct {
	RequestID   *string `json:"requestId"`
	RequestDate *string `json:"requestDate"`
}

// Order represents a payment order.
type Order struct {
	ID                       int                      `json:"id"`
	OrderID                  string                   `json:"orderId"`
	Status                   OrderStatus              `json:"status"`
	FiatAmount               float64                  `json:"fiatAmount"`
	FiatCurrency             FiatCurrency             `json:"fiatCurrency"`
	CryptoAmount             float64                  `json:"cryptoAmount"`
	CryptoCurrency           CryptoCurrency           `json:"cryptoCurrency"`
	ExchangeRate             float64                  `json:"exchangeRate"`
	QRInfo                   QRInfo                   `json:"qrInfo"`
	CryptoTransferInfo       map[string]any           `json:"cryptoTransferInfo"`
	UserID                   *int                     `json:"userId"`
	PaymentMethod            string                   `json:"paymentMethod"`
	TransactionHash          *string                  `json:"transactionHash"`
	WalletAddress            *string                  `json:"walletAddress"`
	Network                  *string                  `json:"network"`
	ExpiresAt                time.Time                `json:"expiresAt"`
	BankTransactionReference BankTransactionReference `json:"bankTransactionReference"`
	Metadata                 map[string]any           `json:"metadata"`
	IsPrefunded              bool                     `json:"isPrefunded"`
	TransactionReference     *string                  `json:"transactionReference"`
	RouteID                  *Route                   `json:"routeId"`
	CreatedAt                time.Time                `json:"createdAt"`
	UpdatedAt                time.Time                `json:"updatedAt"`
}

// Pagination carries paging metadata returned by list endpoints.
type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

// OrderList is the response shape for order list endpoints.
type OrderList struct {
	Items      []Order    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// ExchangeInfo holds the result of an exchange rate calculation.
type ExchangeInfo struct {
	FiatAmount     float64        `json:"fiatAmount"`
	FiatCurrency   string         `json:"fiatCurrency"`
	CryptoAmount   string         `json:"cryptoAmount"`
	CryptoCurrency CryptoCurrency `json:"cryptoCurrency"`
	ExchangeRate   string         `json:"exchangeRate"`
	Chain          Chain          `json:"chain"`
	Token          CryptoCurrency `json:"token"`
	FeeAmount      string         `json:"feeAmount"`
	Timestamp      time.Time      `json:"timestamp"`
}

// MarketInfo describes the payment status for a user in a specific market.
type MarketInfo struct {
	Status       MarketStatus `json:"status"`
	RejectReason *string      `json:"rejectReason"`
	Action       *string      `json:"action"`
}

// PolicyResult is the response from the checkPolicy endpoint.
type PolicyResult struct {
	IsAllowed bool       `json:"isAllowed"`
	RouteID   Route      `json:"routeId"`
	Limits    PolicyLimits `json:"limits"`
	Tier      PolicyTier `json:"tier"`
	Reason    *string    `json:"reason"`
}

// PolicyLimits holds per-transaction and per-day USD limits.
type PolicyLimits struct {
	PerTransaction float64 `json:"perTransaction"`
	PerDay         float64 `json:"perDay"`
}

// TenantBalance holds the prefund balance for the authenticated tenant.
type TenantBalance struct {
	Currency         string  `json:"currency"`
	AvailableBalance float64 `json:"availableBalance"`
	WalletAddress    string  `json:"walletAddress"`
	Chain            Chain   `json:"chain"`
}

// TenantSpend holds the total spent and cap for the authenticated tenant.
type TenantSpend struct {
	TotalSpent float64 `json:"totalSpent"`
	Limit      float64 `json:"limit"`
}
