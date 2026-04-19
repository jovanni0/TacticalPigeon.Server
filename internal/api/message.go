package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type MessageHandler struct {
	DB *sql.DB
}

func (h *MessageHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	chats, err := queries.GetChats(h.DB, userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	partnerID := r.URL.Query().Get("partner_id")
	if partnerID == "" {
		http.Error(w, "missing partner_id", http.StatusBadRequest)
		return
	}
	messages, err := queries.GetMessages(h.DB, userID, partnerID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
