package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/mail"
	"strings"

	dbmodel "newsletter-backend/internal/db"
)

type subscriberInput struct {
	Email string `json:"email"`
}

func (h *Handler) RegisterSubscriberRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscribe", h.createSubscriber)
	mux.HandleFunc("GET /subscribers", h.listSubscribers)
}

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
	writeJSON(w, http.StatusCreated, subscriber)
}

func (h *Handler) listSubscribers(w http.ResponseWriter, r *http.Request) {
	var subscribers []dbmodel.Subscriber
	if err := h.DB.WithContext(r.Context()).Order("subscribed_at DESC").Find(&subscribers).Error; err != nil {
		writeDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, subscribers)
}
