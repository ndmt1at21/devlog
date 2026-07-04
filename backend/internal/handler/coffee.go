package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
)

// allowedCoffeeAmounts are the fixed donation presets (VND). The amount is
// validated server-side so the client can't request an arbitrary charge.
var allowedCoffeeAmounts = map[int64]bool{25000: true, 75000: true, 125000: true}

const coffeeProductName = "Cà phê ủng hộ devnote ☕"

// coffeeCheckout starts a donation. It validates the amount, records a pending
// order, and hands back what the frontend needs: a Stripe redirect URL (card)
// or a MoMo QR (momo). When the chosen provider isn't configured it returns
// {demo:true} so the modal keeps the no-charge demo flow.
func (a *API) coffeeCheckout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Amount int64  `json:"amount"`
		Method string `json:"method"`
		Name   string `json:"name"`
		Email  string `json:"email"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !allowedCoffeeAmounts[in.Amount] {
		writeError(w, r, apierr.ErrCoffeeAmount)
		return
	}
	if in.Method != "card" && in.Method != "momo" {
		writeError(w, r, apierr.ErrCoffeeMethod)
		return
	}

	userID := ""
	if u, ok := userFrom(r.Context()); ok {
		userID = u.Sub
	}

	order := domain.CoffeeOrder{
		ID:         id.NewV7(),
		Method:     in.Method,
		Amount:     in.Amount,
		Currency:   "VND",
		Status:     domain.CoffeePending,
		BuyerName:  strings.TrimSpace(in.Name),
		BuyerEmail: strings.TrimSpace(in.Email),
		UserID:     userID,
	}

	switch in.Method {
	case "card":
		if a.Stripe == nil {
			writeJSON(w, r, http.StatusOK, map[string]any{"demo": true})
			return
		}
		base := strings.TrimRight(a.Cfg.AppBaseURL, "/")
		// {CHECKOUT_SESSION_ID} must remain a literal placeholder for Stripe.
		successURL := fmt.Sprintf("%s/coffee/result?provider=stripe&order=%s&session_id={CHECKOUT_SESSION_ID}", base, order.ID)
		cancelURL := fmt.Sprintf("%s/coffee/result?provider=stripe&order=%s&canceled=1", base, order.ID)
		sessionID, redirectURL, err := a.Stripe.CreateCheckoutSession(r.Context(), payment.CheckoutInput{
			OrderID:     order.ID,
			Amount:      order.Amount,
			ProductName: coffeeProductName,
			Email:       order.BuyerEmail,
			SuccessURL:  successURL,
			CancelURL:   cancelURL,
		})
		if err != nil {
			logger.From(r.Context()).Error("stripe checkout", "err", err.Error(), "order", order.ID)
			writeError(w, r, apierr.ErrCoffeeCheckoutCard)
			return
		}
		order.StripeSessionID = sessionID
		if _, err := a.Store.CoffeeOrders().Create(r.Context(), order); err != nil {
			writeError(w, r, apierr.ErrCoffeeOrderCreate)
			return
		}
		writeJSON(w, r, http.StatusOK, map[string]any{"orderId": order.ID, "redirectUrl": redirectURL})

	case "momo":
		if a.Momo == nil {
			writeJSON(w, r, http.StatusOK, map[string]any{"demo": true})
			return
		}
		requestID := id.NewV7()
		base := strings.TrimRight(a.Cfg.AppBaseURL, "/")
		res, err := a.Momo.CreateOrder(r.Context(), payment.MomoOrderInput{
			OrderID:     order.ID,
			RequestID:   requestID,
			Amount:      order.Amount,
			OrderInfo:   coffeeProductName,
			RedirectURL: fmt.Sprintf("%s/coffee/result?provider=momo&order=%s", base, order.ID),
			IpnURL:      strings.TrimRight(publicBackendURL(r), "/") + apiV1 + "/webhooks/momo",
		})
		if err != nil {
			logger.From(r.Context()).Error("momo create", "err", err.Error(), "order", order.ID)
			writeError(w, r, apierr.ErrCoffeeCheckoutMomo)
			return
		}
		order.MomoOrderID = order.ID
		order.MomoRequestID = requestID
		if _, err := a.Store.CoffeeOrders().Create(r.Context(), order); err != nil {
			writeError(w, r, apierr.ErrCoffeeOrderCreate)
			return
		}
		writeJSON(w, r, http.StatusOK, map[string]any{
			"orderId":   order.ID,
			"qrCodeUrl": res.QRCodeURL,
			"deeplink":  res.Deeplink,
			"payUrl":    res.PayURL,
		})
	}
}

// coffeeStatus returns an order's status, lazily reconciling a still-pending
// order with the provider (Stripe retrieve / MoMo query) so completion is
// confirmed even without a reachable webhook (e.g. local dev). The frontend
// polls this for the MoMo QR flow and the Stripe result page reads it once.
func (a *API) coffeeStatus(w http.ResponseWriter, r *http.Request) {
	oid := r.PathValue("id")
	order, err := a.Store.CoffeeOrders().GetByID(r.Context(), oid)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, r, apierr.ErrCoffeeOrderNotFound)
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrCoffeeLoad)
		return
	}

	if order.Status == domain.CoffeePending {
		if next := a.reconcile(r, order); next != "" && next != order.Status {
			if err := a.Store.CoffeeOrders().UpdateStatus(r.Context(), order.ID, next); err != nil {
				logger.From(r.Context()).Error("coffee update status", "err", err.Error(), "order", order.ID)
			} else {
				order.Status = next
			}
		}
	}

	writeJSON(w, r, http.StatusOK, map[string]any{
		"status": order.Status,
		"amount": order.Amount,
		"method": order.Method,
	})
}

// reconcile asks the provider for the current outcome of a pending order and
// returns the new status, or "" to leave it unchanged.
func (a *API) reconcile(r *http.Request, order domain.CoffeeOrder) string {
	switch order.Method {
	case "card":
		if a.Stripe == nil || order.StripeSessionID == "" {
			return ""
		}
		paid, err := a.Stripe.SessionPaid(r.Context(), order.StripeSessionID)
		if err != nil {
			logger.From(r.Context()).Warn("stripe reconcile", "err", err.Error(), "order", order.ID)
			return ""
		}
		if paid {
			return domain.CoffeeCompleted
		}
	case "momo":
		if a.Momo == nil {
			return ""
		}
		st, err := a.Momo.QueryStatus(r.Context(), order.MomoOrderID, order.MomoRequestID)
		if err != nil {
			logger.From(r.Context()).Warn("momo reconcile", "err", err.Error(), "order", order.ID)
			return ""
		}
		return st // payment.Status* values match domain.Coffee* values
	}
	return ""
}

// stripeWebhook marks an order completed on checkout.session.completed. Stripe
// signs the raw body, so it must be read verbatim before any decoding.
func (a *API) stripeWebhook(w http.ResponseWriter, r *http.Request) {
	if a.Stripe == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	eventType, orderID, err := a.Stripe.VerifyWebhook(body, r.Header.Get("Stripe-Signature"))
	if err != nil {
		logger.From(r.Context()).Warn("stripe webhook rejected", "err", err.Error())
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}
	if eventType == "checkout.session.completed" && orderID != "" {
		if err := a.Store.CoffeeOrders().UpdateStatus(r.Context(), orderID, domain.CoffeeCompleted); err != nil {
			logger.From(r.Context()).Error("stripe webhook update", "err", err.Error(), "order", orderID)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// momoWebhook (IPN) marks an order completed/failed from the verified callback.
func (a *API) momoWebhook(w http.ResponseWriter, r *http.Request) {
	if a.Momo == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ipn, ok, err := a.Momo.VerifyIPN(body)
	if err != nil || !ok {
		logger.From(r.Context()).Warn("momo webhook rejected", "err", err, "ok", ok)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}
	status := domain.CoffeeFailed
	if ipn.ResultCode == 0 {
		status = domain.CoffeeCompleted
	}
	if err := a.Store.CoffeeOrders().UpdateStatus(r.Context(), ipn.OrderID, status); err != nil {
		logger.From(r.Context()).Error("momo webhook update", "err", err.Error(), "order", ipn.OrderID)
	}
	// MoMo expects a 204 acknowledgement.
	w.WriteHeader(http.StatusNoContent)
}

// publicBackendURL derives the externally reachable base URL of this backend
// from the incoming request, honoring reverse-proxy headers. Used for the MoMo
// IPN callback URL.
func publicBackendURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		scheme = p
	}
	host := r.Host
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}
	return scheme + "://" + host
}
