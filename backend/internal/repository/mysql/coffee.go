package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
)

type coffeeRepo struct{ db *sql.DB }

func (r *coffeeRepo) Create(ctx context.Context, o domain.CoffeeOrder) (domain.CoffeeOrder, error) {
	if o.ID == "" {
		o.ID = id.NewV7()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	if o.Status == "" {
		o.Status = domain.CoffeePending
	}
	if _, err := r.db.ExecContext(ctx,
		`INSERT INTO coffee_orders
		 (id, method, amount, currency, status, buyer_name, buyer_email, user_id,
		  stripe_session_id, momo_order_id, momo_request_id, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.Method, o.Amount, o.Currency, o.Status,
		nullStr(o.BuyerName), nullStr(o.BuyerEmail), nullStr(o.UserID),
		nullStr(o.StripeSessionID), nullStr(o.MomoOrderID), nullStr(o.MomoRequestID), o.CreatedAt,
	); err != nil {
		return domain.CoffeeOrder{}, mapError(err)
	}
	return o, nil
}

func (r *coffeeRepo) GetByID(ctx context.Context, oid string) (domain.CoffeeOrder, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, method, amount, currency, status, buyer_name, buyer_email, user_id,
		        stripe_session_id, momo_order_id, momo_request_id, created_at, completed_at
		 FROM coffee_orders WHERE id = ?`, oid)
	var o domain.CoffeeOrder
	var name, email, userID, stripeID, momoOrder, momoReq sql.NullString
	var completed sql.NullTime
	if err := row.Scan(&o.ID, &o.Method, &o.Amount, &o.Currency, &o.Status,
		&name, &email, &userID, &stripeID, &momoOrder, &momoReq, &o.CreatedAt, &completed); err != nil {
		return domain.CoffeeOrder{}, mapError(err)
	}
	o.BuyerName, o.BuyerEmail, o.UserID = name.String, email.String, userID.String
	o.StripeSessionID, o.MomoOrderID, o.MomoRequestID = stripeID.String, momoOrder.String, momoReq.String
	if completed.Valid {
		t := completed.Time
		o.CompletedAt = &t
	}
	return o, nil
}

func (r *coffeeRepo) UpdateStatus(ctx context.Context, oid, status string) error {
	var completedAt any
	if status == domain.CoffeeCompleted {
		completedAt = time.Now().UTC()
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE coffee_orders SET status = ?,
		        completed_at = COALESCE(completed_at, ?)
		 WHERE id = ?`, status, completedAt, oid)
	if err != nil {
		return mapError(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
