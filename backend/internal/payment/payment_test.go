package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

// hmacHex mirrors what the providers compute so tests can build valid payloads.
func hmacHex(secret, msg string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestStripeVerifyWebhook(t *testing.T) {
	const secret = "whsec_test"
	s := NewStripe("sk_test", secret)
	payload := []byte(`{"type":"checkout.session.completed","data":{"object":{"client_reference_id":"order-123","metadata":{}}}}`)
	ts := "1700000000"
	sig := hmacHex(secret, ts+"."+string(payload))

	t.Run("valid", func(t *testing.T) {
		evt, orderID, err := s.VerifyWebhook(payload, fmt.Sprintf("t=%s,v1=%s", ts, sig))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if evt != "checkout.session.completed" {
			t.Errorf("event = %q", evt)
		}
		if orderID != "order-123" {
			t.Errorf("orderID = %q", orderID)
		}
	})

	t.Run("tampered", func(t *testing.T) {
		_, _, err := s.VerifyWebhook(payload, fmt.Sprintf("t=%s,v1=%s", ts, "deadbeef"))
		if err == nil {
			t.Fatal("expected signature mismatch error")
		}
	})

	t.Run("malformed header", func(t *testing.T) {
		if _, _, err := s.VerifyWebhook(payload, "garbage"); err == nil {
			t.Fatal("expected error for malformed header")
		}
	})
}

func TestMomoVerifyIPN(t *testing.T) {
	const accessKey, secret = "ak", "sk"
	m := NewMomo("MOMO", accessKey, secret, "", "")

	fields := map[string]any{
		"partnerCode":  "MOMO",
		"orderId":      "order-9",
		"requestId":    "req-9",
		"amount":       int64(75000),
		"orderInfo":    "coffee",
		"orderType":    "momo_wallet",
		"transId":      int64(123),
		"resultCode":   0,
		"message":      "success",
		"payType":      "qr",
		"responseTime": int64(1700000000000),
		"extraData":    "",
	}
	raw := fmt.Sprintf(
		"accessKey=%s&amount=%d&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%d&resultCode=%d&transId=%d",
		accessKey, fields["amount"], fields["extraData"], fields["message"], fields["orderId"], fields["orderInfo"],
		fields["orderType"], fields["partnerCode"], fields["payType"], fields["requestId"], fields["responseTime"],
		fields["resultCode"], fields["transId"],
	)
	fields["signature"] = hmacHex(secret, raw)
	body, _ := json.Marshal(fields)

	ipn, ok, err := m.VerifyIPN(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected valid signature")
	}
	if ipn.OrderID != "order-9" || ipn.ResultCode != 0 {
		t.Errorf("ipn = %+v", ipn)
	}

	// Tamper the signature → must be rejected.
	fields["signature"] = "bad"
	bad, _ := json.Marshal(fields)
	if _, ok, _ := m.VerifyIPN(bad); ok {
		t.Fatal("expected invalid signature to be rejected")
	}
}
