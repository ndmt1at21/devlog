package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Donation status strings returned by the MoMo client. They match the values of
// the domain.Coffee* status constants.
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// Momo is a minimal MoMo (Vietnam) client for the AIO/captureWallet flow.
type Momo struct {
	httpc          *http.Client
	partnerCode    string
	accessKey      string
	secretKey      string
	createEndpoint string
	queryEndpoint  string
}

// NewMomo builds a MoMo client from sandbox/live credentials and endpoints.
func NewMomo(partnerCode, accessKey, secretKey, createEndpoint, queryEndpoint string) *Momo {
	return &Momo{
		httpc:          &http.Client{Timeout: 15 * time.Second},
		partnerCode:    partnerCode,
		accessKey:      accessKey,
		secretKey:      secretKey,
		createEndpoint: createEndpoint,
		queryEndpoint:  queryEndpoint,
	}
}

func (m *Momo) sign(raw string) string {
	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}

// MomoOrderInput describes a coffee donation to create with MoMo.
type MomoOrderInput struct {
	OrderID     string
	RequestID   string
	Amount      int64 // VND
	OrderInfo   string
	RedirectURL string
	IpnURL      string
}

// MomoOrderResult holds the payment artefacts returned by MoMo.
type MomoOrderResult struct {
	PayURL    string
	Deeplink  string
	QRCodeURL string
}

// CreateOrder creates a captureWallet order and returns the QR/pay artefacts.
func (m *Momo) CreateOrder(ctx context.Context, in MomoOrderInput) (MomoOrderResult, error) {
	const requestType = "captureWallet"
	const extraData = ""
	raw := fmt.Sprintf(
		"accessKey=%s&amount=%d&extraData=%s&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=%s",
		m.accessKey, in.Amount, extraData, in.IpnURL, in.OrderID, in.OrderInfo, m.partnerCode, in.RedirectURL, in.RequestID, requestType,
	)
	payload := map[string]any{
		"partnerCode": m.partnerCode,
		"partnerName": "devnote",
		"storeId":     "devnote",
		"requestId":   in.RequestID,
		"amount":      in.Amount,
		"orderId":     in.OrderID,
		"orderInfo":   in.OrderInfo,
		"redirectUrl": in.RedirectURL,
		"ipnUrl":      in.IpnURL,
		"lang":        "vi",
		"requestType": requestType,
		"extraData":   extraData,
		"signature":   m.sign(raw),
	}
	var out struct {
		ResultCode int    `json:"resultCode"`
		Message    string `json:"message"`
		PayURL     string `json:"payUrl"`
		Deeplink   string `json:"deeplink"`
		QRCodeURL  string `json:"qrCodeUrl"`
	}
	if err := m.postJSON(ctx, m.createEndpoint, payload, &out); err != nil {
		return MomoOrderResult{}, err
	}
	if out.ResultCode != 0 {
		return MomoOrderResult{}, fmt.Errorf("momo create resultCode %d: %s", out.ResultCode, out.Message)
	}
	return MomoOrderResult{PayURL: out.PayURL, Deeplink: out.Deeplink, QRCodeURL: out.QRCodeURL}, nil
}

// QueryStatus polls MoMo for a transaction's outcome and maps the resultCode to
// a donation status (completed | pending | failed). This lets the backend
// confirm payment without a public IPN URL (useful in local dev).
func (m *Momo) QueryStatus(ctx context.Context, orderID, requestID string) (string, error) {
	raw := fmt.Sprintf("accessKey=%s&orderId=%s&partnerCode=%s&requestId=%s",
		m.accessKey, orderID, m.partnerCode, requestID)
	payload := map[string]any{
		"partnerCode": m.partnerCode,
		"requestId":   requestID,
		"orderId":     orderID,
		"lang":        "vi",
		"signature":   m.sign(raw),
	}
	var out struct {
		ResultCode int    `json:"resultCode"`
		Message    string `json:"message"`
	}
	if err := m.postJSON(ctx, m.queryEndpoint, payload, &out); err != nil {
		return "", err
	}
	switch out.ResultCode {
	case 0:
		return StatusCompleted, nil
	case 1000: // initiated, awaiting user confirmation
		return StatusPending, nil
	default:
		return StatusFailed, nil
	}
}

// MomoIPN is the subset of an IPN callback the handler needs.
type MomoIPN struct {
	OrderID    string
	RequestID  string
	ResultCode int
}

// VerifyIPN recomputes the IPN signature and reports whether it is valid.
func (m *Momo) VerifyIPN(body []byte) (MomoIPN, bool, error) {
	var p struct {
		PartnerCode  string `json:"partnerCode"`
		OrderID      string `json:"orderId"`
		RequestID    string `json:"requestId"`
		Amount       int64  `json:"amount"`
		OrderInfo    string `json:"orderInfo"`
		OrderType    string `json:"orderType"`
		TransID      int64  `json:"transId"`
		ResultCode   int    `json:"resultCode"`
		Message      string `json:"message"`
		PayType      string `json:"payType"`
		ResponseTime int64  `json:"responseTime"`
		ExtraData    string `json:"extraData"`
		Signature    string `json:"signature"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return MomoIPN{}, false, err
	}
	raw := fmt.Sprintf(
		"accessKey=%s&amount=%d&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%d&resultCode=%d&transId=%d",
		m.accessKey, p.Amount, p.ExtraData, p.Message, p.OrderID, p.OrderInfo, p.OrderType, p.PartnerCode, p.PayType, p.RequestID, p.ResponseTime, p.ResultCode, p.TransID,
	)
	ok := hmac.Equal([]byte(m.sign(raw)), []byte(p.Signature))
	return MomoIPN{OrderID: p.OrderID, RequestID: p.RequestID, ResultCode: p.ResultCode}, ok, nil
}

func (m *Momo) postJSON(ctx context.Context, endpoint string, payload, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("momo request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("momo status %d: %s", resp.StatusCode, string(respBody))
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("momo decode: %w", err)
	}
	return nil
}
