package ws

import "sync"

// Client represents one connected user.
type Client struct {
	UserID string
	Send   chan []byte // messages queued for this client
}

// Hub holds all connected clients and routes messages between them.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // userID → client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.UserID] = client
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if client, ok := h.clients[userID]; ok {
		close(client.Send)
		delete(h.clients, userID)
	}
}

// SendToUser delivers a message to a specific user if they are connected.
func (h *Hub) SendToUser(userID string, message []byte) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- message:
		default:
			// channel full — client is too slow, drop and disconnect
			h.Unregister(userID)
		}
	}
}

func (h *Hub) isOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
