package main

import (
	"log"
	"net/http"

	"github.com/jovanni0/TacticalPigeon.Server/internal/api"
	"github.com/jovanni0/TacticalPigeon.Server/internal/db"
	"github.com/jovanni0/TacticalPigeon.Server/internal/ws"
)

func main() {
	database, err := db.Open("/data/server.db")

	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}
	defer database.Close()

	inviteHandler := &api.InviteHandler{DB: database}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /admin/invites", inviteHandler.CreateInvite)
	mux.HandleFunc("GET /invites/validate", inviteHandler.ValidateInvite)

	registerHandler := &api.RegisterHandler{DB: database}
	mux.HandleFunc("POST /register", registerHandler.Register)

	authHandler := &api.AuthHandler{DB: database}
	mux.HandleFunc("POST /auth/challenge", authHandler.RequestChallenge)
	mux.HandleFunc("POST /auth/verify", authHandler.VerifyChallenge)

	// example of a protected route — wrap it with RequireAuth
	// mux.HandleFunc("POST /admin/invites", api.RequireAuth(database, inviteHandler.CreateInvite))

	/*
	* handle server check pings from frontend apps.
	 */
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /auth/lookup", authHandler.LookupUser)

	userHandler := &api.UserHandler{DB: database}
	messageHandler := &api.MessageHandler{DB: database}

	hub := ws.NewHub()

	mux.HandleFunc("GET /ws", ws.ServeWS(hub, database))
	mux.HandleFunc("GET /users", api.RequireAuth(database, userHandler.ListUsers))
	mux.HandleFunc("GET /chats", api.RequireAuth(database, messageHandler.GetChats))
	mux.HandleFunc("GET /messages", api.RequireAuth(database, messageHandler.GetMessages))

	mux.HandleFunc("GET /auth/validate", api.RequireAuth(database, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
