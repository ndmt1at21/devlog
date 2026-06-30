package domain

import "time"

// Coffee order statuses.
const (
	CoffeePending   = "pending"
	CoffeeCompleted = "completed"
	CoffeeFailed    = "failed"
	CoffeeCancelled = "cancelled"
)

// CoffeeOrder records a "buy me a coffee" donation. Orders may be anonymous
// (UserID empty). Amount is in VND (zero-decimal). The provider id columns hold
// the Stripe Checkout Session id or the MoMo order/request ids so a payment can
// be reconciled later by querying the provider or via a webhook.
type CoffeeOrder struct {
	ID       string `json:"id"`
	Method   string `json:"method"`   // "card" (Stripe) | "momo"
	Amount   int64  `json:"amount"`   // VND, zero-decimal
	Currency string `json:"currency"` // "VND"
	Status   string `json:"status"`   // pending | completed | failed | cancelled

	BuyerName  string `json:"buyerName,omitempty"`
	BuyerEmail string `json:"buyerEmail,omitempty"`
	UserID     string `json:"-"` // IAM sub when logged in, else empty

	StripeSessionID string `json:"-"`
	MomoOrderID     string `json:"-"` // our order id sent to MoMo (== ID)
	MomoRequestID   string `json:"-"`

	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}
