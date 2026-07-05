// Package upload presigns direct-to-bucket PUTs against an S3-compatible
// object store (Cloudflare R2 in production, MinIO for local dev). Image bytes
// go browser → bucket, never through the API — the backend only authorizes and
// signs. SigV4 query presigning is implemented here with the standard library,
// keeping the backend's single-dependency footprint.
package upload

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signer presigns requests for one bucket on an S3-compatible endpoint,
// addressed path-style (endpoint/bucket/key) as R2 and MinIO expect.
type Signer struct {
	Endpoint  string // scheme://host, e.g. https://<account>.r2.cloudflarestorage.com
	Bucket    string
	Region    string // "auto" on R2
	AccessKey string
	SecretKey string
}

// PresignPut returns a URL that lets its holder PUT exactly one object: the
// key and content type are signed, so neither can be swapped after the
// caller's validation. The client must send a matching Content-Type header.
//
// Content-Length is deliberately NOT signed. R2 verifies a signed
// content-length header inconsistently — signing it makes browser PUTs fail
// with 403/SignatureDoesNotMatch even when the byte count matches. Size is
// already validated server-side before signing, and the object key is
// server-generated and single-use, so leaving Content-Length unsigned only
// loosens the byte-exact cap at the bucket, not the authorization gate.
func (s Signer) PresignPut(key, contentType string, size int64, expires time.Duration, now time.Time) (string, error) {
	u, err := url.Parse(s.Endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid S3 endpoint %q", s.Endpoint)
	}
	path := "/" + s.Bucket + "/" + key
	query := s.presign("PUT", u.Host, path, [][2]string{
		{"content-type", contentType},
	}, expires, now)
	return u.Scheme + "://" + u.Host + escapePath(path) + "?" + query, nil
}

// presign runs the SigV4 query-presigning algorithm and returns the full query
// string, signature included. extraHeaders are headers the client must send
// verbatim; host is always signed.
func (s Signer) presign(method, host, path string, extraHeaders [][2]string, expires time.Duration, now time.Time) string {
	amzDate := now.UTC().Format("20060102T150405Z")
	scope := amzDate[:8] + "/" + s.Region + "/s3/aws4_request"

	headers := append([][2]string{{"host", host}}, extraHeaders...)
	sort.Slice(headers, func(i, j int) bool { return headers[i][0] < headers[j][0] })
	names := make([]string, len(headers))
	var canonHeaders strings.Builder
	for i, h := range headers {
		names[i] = h[0]
		canonHeaders.WriteString(h[0] + ":" + h[1] + "\n")
	}
	signedHeaders := strings.Join(names, ";")

	query := canonicalQuery([][2]string{
		{"X-Amz-Algorithm", "AWS4-HMAC-SHA256"},
		{"X-Amz-Credential", s.AccessKey + "/" + scope},
		{"X-Amz-Date", amzDate},
		{"X-Amz-Expires", strconv.Itoa(int(expires / time.Second))},
		{"X-Amz-SignedHeaders", signedHeaders},
	})

	canonReq := strings.Join([]string{
		method,
		escapePath(path),
		query,
		canonHeaders.String(),
		signedHeaders,
		"UNSIGNED-PAYLOAD",
	}, "\n")
	toSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hexSHA256(canonReq),
	}, "\n")

	k := hmacSHA256([]byte("AWS4"+s.SecretKey), amzDate[:8])
	for _, part := range []string{s.Region, "s3", "aws4_request"} {
		k = hmacSHA256(k, part)
	}
	return query + "&X-Amz-Signature=" + hex.EncodeToString(hmacSHA256(k, toSign))
}

// canonicalQuery encodes and sorts query parameters per the SigV4 canonical
// form (RFC 3986 strict encoding, byte-order sort on encoded keys).
func canonicalQuery(params [][2]string) string {
	pairs := make([]string, len(params))
	for i, p := range params {
		pairs[i] = awsEscape(p[0]) + "=" + awsEscape(p[1])
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "&")
}

// escapePath percent-encodes each path segment, preserving the "/" separators
// (S3 canonical URIs are single-encoded).
func escapePath(path string) string {
	segs := strings.Split(path, "/")
	for i, s := range segs {
		segs[i] = awsEscape(s)
	}
	return strings.Join(segs, "/")
}

// awsEscape percent-encodes everything outside RFC 3986 unreserved characters,
// with uppercase hex — stricter than url.QueryEscape (which keeps "+" for
// spaces and more).
func awsEscape(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func hexSHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key []byte, msg string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return m.Sum(nil)
}
