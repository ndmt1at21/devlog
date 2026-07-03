// Package payment wraps the Stripe and MoMo HTTP APIs used by the "buy me a
// coffee" flow. It uses the raw REST APIs (no SDK dependency). Clients are
// only constructed when the corresponding provider is configured (see
// config.StripeEnabled/MomoEnabled).
package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Stripe is a minimal Stripe client for one-time Checkout payments.
type Stripe struct {
	httpc         *http.Client
	secretKey     string
	webhookSecret string
}

// NewStripe builds a Stripe client. secretKey is the test/live secret key;
// webhookSecret verifies webhook signatures.
func NewStripe(secretKey, webhookSecret string) *Stripe {
	return &Stripe{
		httpc:         &http.Client{Timeout: 15 * time.Second},
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// CheckoutInput describes a one-time coffee donation Checkout Session.
type CheckoutInput struct {
	OrderID     string
	Amount      int64 // VND, zero-decimal (passed to Stripe as-is)
	ProductName string
	Email       string
	SuccessURL  string
	CancelURL   string
}

// CreateCheckoutSession creates a Checkout Session and returns its id and the
// hosted payment URL to redirect the browser to.
func (s *Stripe) CreateCheckoutSession(ctx context.Context, in CheckoutInput) (sessionID, redirectURL string, err error) {
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", in.SuccessURL)
	form.Set("cancel_url", in.CancelURL)
	form.Set("client_reference_id", in.OrderID)
	form.Set("metadata[order_id]", in.OrderID)
	if in.Email != "" {
		form.Set("customer_email", in.Email)
	}
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", "vnd")
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(in.Amount, 10))
	form.Set("line_items[0][price_data][product_data][name]", in.ProductName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.stripe.com/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.secretKey, "")

	resp, err := s.httpc.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("stripe checkout: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("stripe checkout status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("stripe decode: %w", err)
	}
	if out.URL == "" {
		return "", "", fmt.Errorf("stripe checkout: empty url")
	}
	return out.ID, out.URL, nil
}

// SessionPaid reports whether the Checkout Session has been paid. Used to
// reconcile an order's status without relying on a webhook (works in local dev).
func (s *Stripe) SessionPaid(ctx context.Context, sessionID string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.stripe.com/v1/checkout/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(s.secretKey, "")
	resp, err := s.httpc.Do(req)
	if err != nil {
		return false, fmt.Errorf("stripe retrieve: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("stripe retrieve status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		PaymentStatus string `json:"payment_status"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false, err
	}
	return out.PaymentStatus == "paid" || out.PaymentStatus == "no_payment_required", nil
}

// VerifyWebhook validates the Stripe-Signature header against the raw payload
// and returns the event type and the donation's order id (from
// client_reference_id / metadata). It implements the v1 scheme:
// signed_payload = "{t}.{payload}", expected = HMAC-SHA256(secret, signed_payload).
func (s *Stripe) VerifyWebhook(payload []byte, sigHeader string) (eventType, orderID string, err error) {
	var ts string
	var sigs []string
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			sigs = append(sigs, kv[1])
		}
	}
	if ts == "" || len(sigs) == 0 {
		return "", "", fmt.Errorf("stripe webhook: malformed signature header")
	}
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(ts + "." + string(payload)))
	expected := hex.EncodeToString(mac.Sum(nil))
	ok := false
	for _, sig := range sigs {
		if hmac.Equal([]byte(sig), []byte(expected)) {
			ok = true
			break
		}
	}
	if !ok {
		return "", "", fmt.Errorf("stripe webhook: signature mismatch")
	}
	var evt struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ClientReferenceID string            `json:"client_reference_id"`
				Metadata          map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &evt); err != nil {
		return "", "", err
	}
	oid := evt.Data.Object.ClientReferenceID
	if oid == "" {
		oid = evt.Data.Object.Metadata["order_id"]
	}
	return evt.Type, oid, nil
}
