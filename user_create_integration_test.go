package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_CreateUser hits the real sandbox environment. It takes no
// input, so unlike the other integration tests in this package it's
// expected to always succeed (no fixture data, no production-only
// restriction) — every run creates a new real sandbox user.
func TestSandbox_CreateUser(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateUser(ctx, &gaian.CreateUserRequest{})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if resp.Data.UserID == "" {
		t.Fatal("CreateUser: expected a non-empty UserID")
	}
}

func TestSandbox_CreateUser_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.CreateUser(ctx, &gaian.CreateUserRequest{}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
