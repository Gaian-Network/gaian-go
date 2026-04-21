// Example: standard payment flow (register → KYC check → parse QR →
// calculate exchange → place order → verify → poll status).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func main() {
	apiKey := os.Getenv("GAIAN_API_KEY")
	if apiKey == "" {
		log.Fatal("GAIAN_API_KEY environment variable required")
	}

	ctx := context.Background()
	c := gaian.New(apiKey, gaian.Sandbox)

	const wallet = "YOUR_WALLET_ADDRESS"

	// 1. Register user (idempotent)
	user, err := c.User.Register(ctx, wallet)
	if err != nil && !gaian.IsConflict(err) {
		log.Fatalf("register: %v", err)
	}
	if user != nil {
		fmt.Printf("user id: %d  kyc: %s\n", user.ID, user.KYCStatus)
	}

	// 2. Check policy before attempting payment
	policy, err := c.Policy.CheckPolicy(ctx, wallet, "VN")
	if err != nil {
		log.Fatalf("check policy: %v", err)
	}
	if !policy.IsAllowed {
		log.Fatalf("payment not allowed: %v", policy.Reason)
	}
	fmt.Printf("tier: %s  perTx limit: %.2f USD\n", policy.Tier, policy.Limits.PerTransaction)

	// 3. Parse the QR code
	const qrString = "PASTE_QR_STRING_HERE"
	qrInfo, err := c.Payment.ParseQR(ctx, qrString)
	if err != nil {
		log.Fatalf("parse qr: %v", err)
	}
	var amount float64
	if qrInfo.Amount != nil {
		amount = *qrInfo.Amount
	}
	var currency string
	if qrInfo.Currency != nil {
		currency = *qrInfo.Currency
	}
	var beneficiary string
	if qrInfo.BeneficiaryName != nil {
		beneficiary = *qrInfo.BeneficiaryName
	}
	fmt.Printf("beneficiary: %s  amount: %.0f %s\n", beneficiary, amount, currency)

	// 4. Calculate exchange rate
	exchange, err := c.Payment.CalculateExchange(ctx, gaian.CalculateExchangeRequest{
		Amount:  amount,
		Country: "VN",
		Chain:   gaian.Solana,
		Token:   gaian.USDC,
	})
	if err != nil {
		log.Fatalf("calculate exchange: %v", err)
	}
	fmt.Printf("need %s %s (rate: %s)\n",
		exchange.CryptoAmount, exchange.CryptoCurrency, exchange.ExchangeRate)

	// 5. Place order
	order, err := c.Payment.PlaceOrder(ctx, gaian.PlaceOrderRequest{
		QRString:       qrString,
		Amount:         amount,
		CryptoCurrency: gaian.USDC,
		FromAddress:    wallet,
		FiatCurrency:   gaian.VND,
		Chain:          gaian.Solana,
	})
	if err != nil {
		log.Fatalf("place order: %v", err)
	}
	fmt.Printf("order %s created, status: %s\n", order.OrderID, order.Status)

	// 6. TODO: sign and broadcast the on-chain Solana transaction using
	//    order.CryptoTransferInfo, then call VerifyOrder with the signature.
	txProof := "TRANSACTION_SIGNATURE_PLACEHOLDER"

	result, err := c.Payment.VerifyOrder(ctx, order.OrderID, txProof)
	if err != nil {
		log.Fatalf("verify order: %v", err)
	}
	fmt.Printf("verify: %s  bank transfer: %v\n", result.Status, result.BankTransferStatus)

	// 7. Poll until terminal state
	for {
		status, err := c.Payment.GetOrderStatus(ctx, order.OrderID)
		if err != nil {
			log.Fatalf("get status: %v", err)
		}
		fmt.Printf("polling status: %s\n", status.Status)
		if status.Status == gaian.OrderCompleted || status.Status == gaian.OrderFailed {
			fmt.Printf("final status: %s\n", status.Status)
			break
		}
		time.Sleep(3 * time.Second)
	}
}
