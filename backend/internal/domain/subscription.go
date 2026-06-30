package domain

import "time"

// Subscription records a (demo) Pro membership, keyed by the IAM user id (sub).
type Subscription struct {
	ID        string    `json:"-"`
	UserID    string    `json:"-"`
	Plan      string    `json:"plan"`   // "month" | "year"
	Status    string    `json:"status"` // "active"
	CreatedAt time.Time `json:"createdAt"`
}
