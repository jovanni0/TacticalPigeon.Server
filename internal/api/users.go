package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type UserHandler struct {
	DB *sql.DB
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := queries.GetAllUsers(h.DB)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
