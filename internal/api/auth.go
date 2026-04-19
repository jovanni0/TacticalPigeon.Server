package api

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type AuthHandler struct {
	DB *sql.DB
}

func (h *AuthHandler) RequestChallenge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	// make sure the user actually exists
	user, err := queries.GetUserByID(h.DB, req.UserID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	challengeToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * time.Second)

	if err := queries.CreateChallenge(h.DB, challengeToken, req.UserID, expiresAt); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"challenge": challengeToken,
	})
}

func (h *AuthHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		Challenge string `json:"challenge"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// look up the challenge
	userID, expiresAt, err := queries.GetChallenge(h.DB, req.Challenge)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// challenge not found, already used, expired, or belongs to a different user
	if userID == "" || time.Now().After(expiresAt) || userID != req.UserID {
		http.Error(w, "invalid or expired challenge", http.StatusUnauthorized)
		return
	}

	// delete the challenge immediately — it's single use
	if err := queries.DeleteChallenge(h.DB, req.Challenge); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// get the user's stored public key
	user, err := queries.GetUserByID(h.DB, req.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// decode the public key from base64
	pubKeyBytes, err := base64.StdEncoding.DecodeString(user.PublicKey)
	if err != nil {
		http.Error(w, "invalid public key on record", http.StatusInternalServerError)
		return
	}

	// decode the signature from base64
	sigBytes, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		http.Error(w, "invalid signature encoding", http.StatusBadRequest)
		return
	}

	// verify — the app signed the challenge string using the private key
	pubKey := ed25519.PublicKey(pubKeyBytes)
	if !ed25519.Verify(pubKey, []byte(req.Challenge), sigBytes) {
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	// signature is valid — create a session
	sessionToken := uuid.NewString()
	sessionExpiry := time.Now().Add(30 * 24 * time.Hour)

	if err := queries.CreateSession(h.DB, sessionToken, req.UserID, sessionExpiry); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_token": sessionToken,
	})
}

func (h *AuthHandler) LookupUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PublicKey == "" {
		http.Error(w, "missing public_key", http.StatusBadRequest)
		return
	}

	user, err := queries.GetUserByPublicKey(h.DB, req.PublicKey)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "no account found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": user.ID,
	})
}
