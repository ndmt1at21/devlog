package handler

import (
	"errors"
	"net/http"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
)

type planDTO struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Price string `json:"price"`
	Note  string `json:"note"`
}

func (a *API) proPlans(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, []planDTO{
		{Key: "month", Name: "Theo tháng", Price: "39K", Note: "thanh toán hàng tháng"},
		{Key: "year", Name: "Theo năm", Price: "299K", Note: "~25K/tháng · tiết kiệm 36%"},
	})
}

func (a *API) getSubscription(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	s, err := a.Store.Subscriptions().GetByUser(r.Context(), u.Sub)
	if errors.Is(err, domain.ErrNotFound) || s == nil {
		writeJSON(w, r, http.StatusOK, map[string]any{"active": false})
		return
	}
	if err != nil {
		writeError(w, r, apierr.ErrSubscriptionLoad)
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]any{"active": s.Status == "active", "plan": s.Plan, "status": s.Status})
}

// subscribe activates a demo Pro membership (no real payment).
func (a *API) subscribe(w http.ResponseWriter, r *http.Request) {
	u, ok := userFrom(r.Context())
	if !ok {
		writeError(w, r, apierr.ErrUnauthorized)
		return
	}
	var in struct {
		Plan string `json:"plan"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Plan != "month" && in.Plan != "year" {
		writeError(w, r, apierr.ErrPlanInvalid)
		return
	}
	s, err := a.Store.Subscriptions().Create(r.Context(), domain.Subscription{
		UserID: u.Sub, Plan: in.Plan, Status: "active",
	})
	if err != nil {
		writeError(w, r, apierr.ErrSubscribeFailed)
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]any{"active": true, "plan": s.Plan, "status": s.Status})
}
