package api

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// expect: "Authorization: Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, expiresAt, err := queries.GetSession(db, token)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if userID == "" || time.Now().After(expiresAt) {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}

		// attach the user ID to the request context so handlers can read it
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
