package upload

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestPresignAWSExampleVector pins the signer to the worked example in the AWS
// SigV4 documentation ("Authenticating Requests: Using Query Parameters"):
// a presigned GET of s3://examplebucket/test.txt with only the host header
// signed. If the canonical request, string-to-sign, or key derivation drift,
// this signature changes.
func TestPresignAWSExampleVector(t *testing.T) {
	s := Signer{
		Endpoint:  "https://examplebucket.s3.amazonaws.com",
		Region:    "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
	now := time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC)
	query := s.presign("GET", "examplebucket.s3.amazonaws.com", "/test.txt", nil, 86400*time.Second, now)

	const wantSig = "aeeed9bbccd4d02ee5c0109b86d86835f995330da4c265957d157751f604d404"
	if !strings.HasSuffix(query, "&X-Amz-Signature="+wantSig) {
		t.Errorf("signature mismatch:\n got  %s\n want …&X-Amz-Signature=%s", query, wantSig)
	}
	const wantQuery = "X-Amz-Algorithm=AWS4-HMAC-SHA256" +
		"&X-Amz-Credential=AKIAIOSFODNN7EXAMPLE%2F20130524%2Fus-east-1%2Fs3%2Faws4_request" +
		"&X-Amz-Date=20130524T000000Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host"
	if !strings.HasPrefix(query, wantQuery) {
		t.Errorf("canonical query mismatch:\n got  %s\n want %s…", query, wantQuery)
	}
}

func TestPresignPutShape(t *testing.T) {
	s := Signer{
		Endpoint:  "https://acc123.r2.cloudflarestorage.com",
		Bucket:    "devlog-images",
		Region:    "auto",
		AccessKey: "ak",
		SecretKey: "sk",
	}
	now := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	raw, err := s.PresignPut("img/0197-abc.png", "image/png", 12345, 10*time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("presigned URL does not parse: %v", err)
	}
	if u.Host != "acc123.r2.cloudflarestorage.com" || u.Path != "/devlog-images/img/0197-abc.png" {
		t.Errorf("host/path = %s %s", u.Host, u.Path)
	}
	q := u.Query()
	if got := q.Get("X-Amz-SignedHeaders"); got != "content-type;host" {
		t.Errorf("SignedHeaders = %q, want content-type;host", got)
	}
	if got := q.Get("X-Amz-Expires"); got != "600" {
		t.Errorf("Expires = %q, want 600", got)
	}
	if got := q.Get("X-Amz-Credential"); !strings.HasPrefix(got, "ak/20260703/auto/s3/aws4_request") {
		t.Errorf("Credential = %q", got)
	}
	if sig := q.Get("X-Amz-Signature"); len(sig) != 64 {
		t.Errorf("signature = %q, want 64 hex chars", sig)
	}
}

func TestPresignPutRejectsBadEndpoint(t *testing.T) {
	if _, err := (Signer{Endpoint: "not a url"}).PresignPut("k", "image/png", 1, time.Minute, time.Now()); err == nil {
		t.Error("want error for invalid endpoint")
	}
}
