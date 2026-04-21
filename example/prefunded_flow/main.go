// Example: prefunded payment flow (no on-chain signing required).
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

	// Check tenant prefund balance first
	balance, err := c.Tenant.GetBalance(ctx, gaian.USDC)
	if err != nil {
		log.Fatalf("get balance: %v", err)
	}
	fmt.Printf("available balance: %.4f %s on %s\n",
		balance.AvailableBalance, balance.Currency, balance.Chain)

	const (
		wallet   = "YOUR_WALLET_ADDRESS"
		qrString = "PASTE_QR_STRING_HERE"
	)

	order, err := c.Payment.PlacePrefundedOrder(ctx, gaian.PlacePrefundedOrderRequest{
		QRString:       qrString,
		Amount:         50000,
		CryptoCurrency: gaian.USDC,
		FromAddress:    wallet,
		FiatCurrency:   gaian.VND,
	})
	if err != nil {
		log.Fatalf("place prefunded order: %v", err)
	}
	fmt.Printf("prefunded order %s created\n", order.OrderID)

	// Poll — no VerifyOrder step needed for prefunded flow
	for {
		status, err := c.Payment.GetOrderStatus(ctx, order.OrderID)
		if err != nil {
			log.Fatalf("get status: %v", err)
		}
		fmt.Printf("status: %s\n", status.Status)
		if status.Status == gaian.OrderCompleted || status.Status == gaian.OrderFailed {
			break
		}
		time.Sleep(3 * time.Second)
	}
}
