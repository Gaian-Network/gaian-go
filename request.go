package gaian

import (
	"io"
	"net/http"
)

// request is the internal representation of a single API call, built up
// by Client.buildRequest and consumed by Client.excute. Endpoint files
// only ever need to set Method, Endpoint, and (when there's a body or
// query to send) Params — every other field is populated internally and
// setting it at the call site has no effect (see the notes on each
// field below).
type request struct {
	// Method is the HTTP method, e.g. http.MethodGet/http.MethodPost.
	Method string

	// Endpoint is the request path, e.g. "/api/v2/users/123" — not a
	// full URL; Client prepends its configured base URL.
	Endpoint string

	// Params is the request body (for non-GET methods) or query
	// parameters (for GET, via paramsToMap). Leave nil when there's
	// nothing to send.
	Params any

	// SigningData is currently unused by the signing path (buildRequest
	// overwrites it from Params, and sign() reads Params directly, not
	// this field) — do not set it expecting it to affect the computed
	// signature.
	SigningData []byte

	// Header is populated by buildRequest; setting it beforehand is
	// pointless — it's cloned/overwritten.
	Header http.Header

	// Body is the request body reader, populated by buildRequest from
	// the marshaled Params.
	Body io.Reader

	// FullURL is the resolved baseURL+Endpoint(+query), populated by
	// buildRequest.
	FullURL string

	// Public, when true, skips HMAC signing entirely. Every current v2
	// endpoint requires signing, so this should stay false unless a
	// future public (unauthenticated) endpoint needs it.
	Public bool
}
