package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type RegisterHandler struct {
	DB *sql.DB
}

type registerRequest struct {
	InviteToken   string `json:"invite_token"`
	DisplayName   string `json:"display_name"`
	ContactPolicy string `json:"contact_policy"`
	PublicKey     string `json:"public_key"`
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// basic validation — all fields are required
	if req.InviteToken == "" || req.DisplayName == "" || req.PublicKey == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	if req.ContactPolicy != "open" && req.ContactPolicy != "approval_required" {
		http.Error(w, "contact_policy must be 'open' or 'approval_required'", http.StatusBadRequest)
		return
	}

	// validate the invite
	inv, err := queries.GetInvite(h.DB, req.InviteToken)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if inv == nil || inv.UsedBy != nil || time.Now().After(inv.ExpirationTime) {
		http.Error(w, "invalid or expired invite", http.StatusNotFound)
		return
	}

	// create the user
	userID := uuid.NewString()
	err = queries.CreateUser(h.DB, req.InviteToken, userID, req.DisplayName, req.ContactPolicy, req.PublicKey)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
	})
}
