package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type InviteHandler struct {
	DB *sql.DB
}

func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	// hardcoded for now — later this comes from the session token
	adminID := "admin"

	token := uuid.NewString()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := queries.CreateInvite(h.DB, token, adminID, expiresAt); err != nil {
		http.Error(w, "failed to create invite", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
		"link":  "https://yourdomain.com/register?invite=" + token,
	})
}

func (h *InviteHandler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	inv, err := queries.GetInvite(h.DB, token)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if inv == nil || inv.UsedBy != nil || time.Now().After(inv.ExpirationTime) {
		http.Error(w, "invalid or expired invite", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
