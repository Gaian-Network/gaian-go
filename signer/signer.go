package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid api key")
	ErrInvalidSecretKey = errors.New("invalid secret key")
)

type SignParams struct {
	Method string
	Path   string
	Query  string
	Body   any
}

func BuildHeaders(apiKey, secretKey string, payload SignParams) (map[string]string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	sortedBody, err := marshalSorted(payload.Body)
	if err != nil {
		return nil, fmt.Errorf("gaian signer: %w", err)
	}

	message := apiKey + payload.Method + payload.Path + payload.Query + sortedBody + timestamp
	signature := Sign(secretKey, message)

	requestID := uuid.New().String()

	headers := map[string]string{
		"X-GAIAN-KEY":       apiKey,
		"X-GAIAN-SIGNATURE": signature,
		"X-GAIAN-TIMESTAMP": timestamp,
		"X-REQUEST-ID":      requestID,
	}

	return headers, nil
}

func marshalSorted(v any) (string, error) {
	if v == nil {
		return "", nil
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return string(raw), err
	}

	return buildSortedBody(body)
}

func buildSortedBody(body map[string]any) (string, error) {
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]byte, 0, 128)
	out = append(out, '{')

	for i, k := range keys {
		kBytes, err := json.Marshal(k)
		if err != nil {
			return "", err
		}

		vBytes, err := json.Marshal(body[k])
		if err != nil {
			return "", err
		}
		out = append(out, kBytes...)
		out = append(out, ':')
		out = append(out, vBytes...)
		if i < len(keys)-1 {
			out = append(out, ',')
		}
	}

	out = append(out, '}')
	return string(out), nil
}

func Sign(secretKey, message string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signature = strings.ReplaceAll(signature, "+", "-")
	signature = strings.ReplaceAll(signature, "/", "_")

	return signature
}
