package ws

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jovanni0/TacticalPigeon.Server/internal/db/queries"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// envelope wraps every WS message in both directions.
type envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sendMessagePayload struct {
	ReceiverID string `json:"receiver_id"`
	Body       string `json:"body"`
}

type markReadPayload struct {
	MessageID string `json:"message_id"`
}

func ServeWS(hub *Hub, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// the session token comes as a query param for WS
		// because you can't set headers on a browser WebSocket.
		// from Flutter you could use headers, but query param works everywhere.
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		userID, expiresAt, err := queries.GetSession(db, token)
		if err != nil || userID == "" || time.Now().After(expiresAt) {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("ws upgrade error:", err)
			return
		}

		client := &Client{UserID: userID, Send: make(chan []byte, 64)}
		hub.Register(client)
		defer hub.Unregister(userID)

		// goroutine: pump outgoing messages to the connection
		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					break
				}
			}
			conn.Close()
		}()

		// mark any undelivered messages as delivered now that user is online
		go markPendingDelivered(db, hub, userID)

		// main loop: read incoming messages from this client
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				break // client disconnected
			}

			var env envelope
			if err := json.Unmarshal(raw, &env); err != nil {
				continue
			}

			switch env.Type {
			case "send_message":
				handleSendMessage(db, hub, userID, env.Payload)
			case "mark_read":
				handleMarkRead(db, hub, userID, env.Payload)
			}
		}
	}
}

func handleSendMessage(db *sql.DB, hub *Hub, senderID string, raw json.RawMessage) {
	var p sendMessagePayload
	if err := json.Unmarshal(raw, &p); err != nil || p.ReceiverID == "" || p.Body == "" {
		return
	}

	msg, err := queries.CreateMessage(db, uuid.NewString(), senderID, p.ReceiverID, p.Body)
	if err != nil {
		log.Println("create message error:", err)
		return
	}

	event, _ := json.Marshal(envelope{
		Type:    "new_message",
		Payload: mustMarshal(msg),
	})

	// deliver to receiver if online — status becomes 'delivered'
	hub.SendToUser(p.ReceiverID, event)
	if hub.isOnline(p.ReceiverID) {
		queries.UpdateMessageStatus(db, msg.ID, "delivered")
		// notify sender the message was delivered
		delivered, _ := json.Marshal(envelope{
			Type:    "message_delivered",
			Payload: mustMarshal(map[string]string{"message_id": msg.ID}),
		})
		hub.SendToUser(senderID, delivered)
	}

	// always echo back to sender so their UI updates
	hub.SendToUser(senderID, event)
}

func handleMarkRead(db *sql.DB, hub *Hub, readerID string, raw json.RawMessage) {
	var p markReadPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.MessageID == "" {
		return
	}

	queries.UpdateMessageStatus(db, p.MessageID, "read")

	// look up who sent this message so we can notify them
	msg, err := queries.GetMessageByID(db, p.MessageID)
	if err != nil || msg == nil {
		return
	}

	event, _ := json.Marshal(envelope{
		Type:    "message_read",
		Payload: mustMarshal(map[string]string{"message_id": p.MessageID}),
	})
	hub.SendToUser(msg.SenderID, event)
}

func markPendingDelivered(db *sql.DB, hub *Hub, userID string) {
	messages, err := queries.GetUndeliveredMessages(db, userID)
	if err != nil {
		return
	}
	for _, msg := range messages {
		queries.UpdateMessageStatus(db, msg.ID, "delivered")
		event, _ := json.Marshal(envelope{
			Type:    "message_delivered",
			Payload: mustMarshal(map[string]string{"message_id": msg.ID}),
		})
		hub.SendToUser(msg.SenderID, event)
	}
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
