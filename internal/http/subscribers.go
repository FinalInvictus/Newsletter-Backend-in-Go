package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/mail"
	"strings"
	"time"

	dbmodel "github.com/FinalInvictus/Newsletter-Backend-in-Go.git/internal/db"
)

type subscriberInput struct {
	Email string `json:"email"`
}

// subscriberView exposes signup details needed by the editor without returning unsubscribe secrets.
type subscriberView struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	SubscribedAt time.Time `json:"subscribed_at"`
	IsVerified   bool      `json:"is_verified"`
}

// RegisterSubscriberRoutes binds newsletter signup and subscriber listing endpoints.
func (h *Handler) RegisterSubscriberRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscribe", h.createSubscriber)
	mux.HandleFunc("GET /subscribers", h.listSubscribers)
}

// createSubscriber validates an email, creates a private unsubscribe token, and stores the signup.
func (h *Handler) createSubscriber(w http.ResponseWriter, r *http.Request) {
	var input subscriberInput
	if !decodeJSON(w, r, &input) {
		return
	}
	email := strings.TrimSpace(input.Email)
	if email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		writeError(w, http.StatusBadRequest, "email must be valid")
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create subscriber token")
		return
	}
	token := hex.EncodeToString(tokenBytes)
	subscriber := dbmodel.Subscriber{Email: strings.ToLower(email), UnsubscribeToken: &token, IsVerified: false}
	if err := h.DB.WithContext(r.Context()).Create(&subscriber).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, subscriberView{ID: subscriber.ID.String(), Email: subscriber.Email, SubscribedAt: subscriber.SubscribedAt, IsVerified: subscriber.IsVerified})
}

// listSubscribers returns the publication's subscriber records in signup order.
func (h *Handler) listSubscribers(w http.ResponseWriter, r *http.Request) {
	var subscribers []dbmodel.Subscriber
	if err := h.DB.WithContext(r.Context()).Order("subscribed_at DESC").Find(&subscribers).Error; err != nil {
		writeDBError(w, err)
		return
	}
	views := make([]subscriberView, 0, len(subscribers))
	for _, subscriber := range subscribers {
		views = append(views, subscriberView{ID: subscriber.ID.String(), Email: subscriber.Email, SubscribedAt: subscriber.SubscribedAt, IsVerified: subscriber.IsVerified})
	}
	writeJSON(w, http.StatusOK, views)
}
