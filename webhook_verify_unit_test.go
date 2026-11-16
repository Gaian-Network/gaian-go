package gaian_test

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

type webhookVector struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
	Signature string `json:"signature"`
}

func loadWebhookVectors(t *testing.T) (string, []webhookVector) {
	t.Helper()

	raw, err := os.ReadFile("testdata/signature-vectors.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		WebhookSecret string          `json:"webhookSecret"`
		Webhooks      []webhookVector `json:"webhooks"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if len(file.Webhooks) == 0 {
		t.Fatal("no webhook vectors in testdata")
	}

	return file.WebhookSecret, file.Webhooks
}

func vectorTime(t *testing.T, v webhookVector) time.Time {
	t.Helper()

	sec, err := strconv.ParseInt(v.Timestamp, 10, 64)
	if err != nil {
		t.Fatalf("vector timestamp: %v", err)
	}

	return time.Unix(sec, 0)
}

func TestVerifyWebhook(t *testing.T) {
	t.Parallel()

	secret, vectors := loadWebhookVectors(t)

	t.Run("accepts every gateway-generated vector", func(t *testing.T) {
		t.Parallel()

		for _, v := range vectors {
			err := gaian.VerifyWebhook(gaian.VerifyWebhookParams{
				Secret: secret, Signature: v.Signature, Timestamp: v.Timestamp, Body: []byte(v.Body), Now: vectorTime(t, v),
			})
			if err != nil {
				t.Errorf("%s: unexpected error: %v", v.Name, err)
			}
		}
	})

	v := vectors[0]
	now := vectorTime(t, v)
	base := func() gaian.VerifyWebhookParams {
		return gaian.VerifyWebhookParams{
			Secret: secret, Signature: v.Signature, Timestamp: v.Timestamp, Body: []byte(v.Body), Now: now,
		}
	}

	rejects := []struct {
		name string
		mut  func(*gaian.VerifyWebhookParams)
		want error
	}{
		{"wrong secret", func(p *gaian.VerifyWebhookParams) { p.Secret = "whsec_other" }, gaian.ErrInvalidWebhookSignature},
		{"body changed by one byte", func(p *gaian.VerifyWebhookParams) { p.Body = []byte(v.Body + " ") }, gaian.ErrInvalidWebhookSignature},
		{"re-serialized body", func(p *gaian.VerifyWebhookParams) { p.Body = []byte(`{"event":"order.completed"}`) }, gaian.ErrInvalidWebhookSignature},
		{"timestamp swapped", func(p *gaian.VerifyWebhookParams) { p.Timestamp = "1731772801" }, gaian.ErrInvalidWebhookSignature},
		{"missing sha256 prefix", func(p *gaian.VerifyWebhookParams) { p.Signature = v.Signature[len("sha256="):] }, gaian.ErrInvalidWebhookSignature},
		{"not hex", func(p *gaian.VerifyWebhookParams) { p.Signature = "sha256=zzzz" }, gaian.ErrInvalidWebhookSignature},
		{"truncated hex", func(p *gaian.VerifyWebhookParams) { p.Signature = v.Signature[:len(v.Signature)-2] }, gaian.ErrInvalidWebhookSignature},
		{"empty signature", func(p *gaian.VerifyWebhookParams) { p.Signature = "" }, gaian.ErrInvalidWebhookSignature},
		{"empty secret", func(p *gaian.VerifyWebhookParams) { p.Secret = "" }, gaian.ErrInvalidWebhookSignature},
		{"non numeric timestamp", func(p *gaian.VerifyWebhookParams) { p.Timestamp = "yesterday" }, gaian.ErrInvalidWebhookSignature},
		{"301 s in the past", func(p *gaian.VerifyWebhookParams) { p.Now = now.Add(301 * time.Second) }, gaian.ErrWebhookTimestampOutOfTolerance},
		{"301 s in the future", func(p *gaian.VerifyWebhookParams) { p.Now = now.Add(-301 * time.Second) }, gaian.ErrWebhookTimestampOutOfTolerance},
		{"outside a custom tolerance", func(p *gaian.VerifyWebhookParams) {
			p.ToleranceSeconds = 10
			p.Now = now.Add(11 * time.Second)
		}, gaian.ErrWebhookTimestampOutOfTolerance},
	}

	for _, tc := range rejects {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			t.Parallel()

			p := base()
			tc.mut(&p)

			if err := gaian.VerifyWebhook(p); !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
		})
	}

	accepts := []struct {
		name string
		mut  func(*gaian.VerifyWebhookParams)
	}{
		{"exactly 300 s old (default tolerance)", func(p *gaian.VerifyWebhookParams) { p.Now = now.Add(300 * time.Second) }},
		{"exactly 300 s ahead (default tolerance)", func(p *gaian.VerifyWebhookParams) { p.Now = now.Add(-300 * time.Second) }},
		{"inside a custom tolerance", func(p *gaian.VerifyWebhookParams) {
			p.ToleranceSeconds = 600
			p.Now = now.Add(599 * time.Second)
		}},
	}

	for _, tc := range accepts {
		t.Run("accepts "+tc.name, func(t *testing.T) {
			t.Parallel()

			p := base()
			tc.mut(&p)

			if err := gaian.VerifyWebhook(p); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// With Now left zero, VerifyWebhook uses the real clock, so the 2024
// vector must be rejected as stale even though its signature is valid.
func TestVerifyWebhook_UsesRealClockByDefault(t *testing.T) {
	t.Parallel()

	secret, vectors := loadWebhookVectors(t)
	v := vectors[0]

	err := gaian.VerifyWebhook(gaian.VerifyWebhookParams{
		Secret: secret, Signature: v.Signature, Timestamp: v.Timestamp, Body: []byte(v.Body),
	})
	if !errors.Is(err, gaian.ErrWebhookTimestampOutOfTolerance) {
		t.Errorf("error = %v, want %v", err, gaian.ErrWebhookTimestampOutOfTolerance)
	}
}

func TestVerifyWebhook_NonPositiveToleranceUsesDefault(t *testing.T) {
	t.Parallel()

	secret, vectors := loadWebhookVectors(t)
	v := vectors[0]
	now := vectorTime(t, v)

	for _, tolerance := range []int{0, -1} {
		p := gaian.VerifyWebhookParams{
			Secret: secret, Signature: v.Signature, Timestamp: v.Timestamp, Body: []byte(v.Body),
			ToleranceSeconds: tolerance, Now: now.Add(300 * time.Second),
		}
		if err := gaian.VerifyWebhook(p); err != nil {
			t.Errorf("tolerance %d at 300 s: unexpected error: %v", tolerance, err)
		}

		p.Now = now.Add(301 * time.Second)
		if err := gaian.VerifyWebhook(p); !errors.Is(err, gaian.ErrWebhookTimestampOutOfTolerance) {
			t.Errorf("tolerance %d at 301 s: error = %v", tolerance, err)
		}
	}
}
