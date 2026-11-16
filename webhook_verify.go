package gaian

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookToleranceSeconds is how far a webhook timestamp may be
// from the current time, in either direction, when
// VerifyWebhookParams.ToleranceSeconds is not set.
const DefaultWebhookToleranceSeconds = 300

const webhookSignaturePrefix = "sha256="

var (
	// ErrInvalidWebhookSignature means the signature, timestamp or secret
	// is missing or malformed, or the signature does not match the body.
	ErrInvalidWebhookSignature = errors.New("gaian: invalid webhook signature")
	// ErrWebhookTimestampOutOfTolerance means the signature is valid but
	// the timestamp is too old or too far in the future; treat the
	// delivery as a replay.
	ErrWebhookTimestampOutOfTolerance = errors.New("gaian: webhook timestamp outside tolerance")
)

// VerifyWebhookParams is the input of VerifyWebhook.
type VerifyWebhookParams struct {
	// Secret is the signing secret of the webhook endpoint.
	Secret string
	// Signature is the X-Gaian-Signature header ("sha256=<hex>").
	Signature string
	// Timestamp is the X-Gaian-Timestamp header (Unix seconds).
	Timestamp string
	// Body is the request body exactly as received. Do not decode and
	// re-encode it: any change of bytes breaks the signature.
	Body []byte
	// ToleranceSeconds is the accepted distance between Timestamp and
	// Now. Zero or negative means DefaultWebhookToleranceSeconds.
	ToleranceSeconds int
	// Now is the current time. The zero value means time.Now().
	Now time.Time
}

// VerifyWebhook checks that a webhook delivery was signed by Gaian with
// the endpoint's secret and is recent. It returns nil when the delivery
// is authentic, ErrInvalidWebhookSignature when the signature does not
// match, and ErrWebhookTimestampOutOfTolerance when it matches but the
// timestamp is outside the tolerance. The comparison is constant-time.
//
// Example:
//
//	body, _ := io.ReadAll(r.Body)
//	err := gaian.VerifyWebhook(gaian.VerifyWebhookParams{
//		Secret:    webhookSecret,
//		Signature: r.Header.Get("X-Gaian-Signature"),
//		Timestamp: r.Header.Get("X-Gaian-Timestamp"),
//		Body:      body,
//	})
//	if err != nil {
//		http.Error(w, "invalid signature", http.StatusUnauthorized)
//		return
//	}
func VerifyWebhook(p VerifyWebhookParams) error {
	if p.Secret == "" {
		return ErrInvalidWebhookSignature
	}

	sentAt, err := strconv.ParseInt(p.Timestamp, 10, 64)
	if err != nil {
		return ErrInvalidWebhookSignature
	}

	hexSignature, ok := strings.CutPrefix(p.Signature, webhookSignaturePrefix)
	if !ok {
		return ErrInvalidWebhookSignature
	}

	got, err := hex.DecodeString(hexSignature)
	if err != nil {
		return ErrInvalidWebhookSignature
	}

	mac := hmac.New(sha256.New, []byte(p.Secret))
	mac.Write([]byte(p.Timestamp))
	mac.Write([]byte("."))
	mac.Write(p.Body)

	if !hmac.Equal(got, mac.Sum(nil)) {
		return ErrInvalidWebhookSignature
	}

	now := p.Now
	if now.IsZero() {
		now = time.Now()
	}

	tolerance := int64(p.ToleranceSeconds)
	if tolerance <= 0 {
		tolerance = DefaultWebhookToleranceSeconds
	}

	age := now.Unix() - sentAt
	if age > tolerance || age < -tolerance {
		return ErrWebhookTimestampOutOfTolerance
	}

	return nil
}
