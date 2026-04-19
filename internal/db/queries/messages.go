package queries

import (
	"database/sql"
	"fmt"
	"time"
)

type Message struct {
	ID           string    `json:"id"`
	SenderID     string    `json:"sender_id"`
	ReceiverID   string    `json:"receiver_id"`
	Body         string    `json:"body"`
	Status       string    `json:"status"`
	CreationDate time.Time `json:"creation_date"`
}

// returns all users the given user has exchanged messages with,
// plus the last message for each — this populates the chats list.
func GetChats(db *sql.DB, userID string) ([]Message, error) {
	rows, err := db.Query(`
        SELECT m.id, m.sender_id, m.receiver_id, m.body, m.status, m.creation_date
        FROM messages m
        INNER JOIN (
            SELECT
                CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner,
                MAX(creation_date) AS latest
            FROM messages
            WHERE sender_id = ? OR receiver_id = ?
            GROUP BY partner
        ) latest ON (
            (m.sender_id = ? AND m.receiver_id = latest.partner) OR
            (m.receiver_id = ? AND m.sender_id = latest.partner)
        ) AND m.creation_date = latest.latest
        ORDER BY m.creation_date DESC`,
		userID, userID, userID, userID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting chats: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var creationDate string
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Body, &msg.Status, &creationDate); err != nil {
			return nil, err
		}
		msg.CreationDate, _ = time.Parse(time.RFC3339, creationDate)
		messages = append(messages, msg)
	}
	return messages, nil
}

// returns full message history between two users.
func GetMessages(db *sql.DB, userA, userB string) ([]Message, error) {
	rows, err := db.Query(`
        SELECT id, sender_id, receiver_id, body, status, creation_date
        FROM messages
        WHERE (sender_id = ? AND receiver_id = ?)
           OR (sender_id = ? AND receiver_id = ?)
        ORDER BY creation_date ASC`,
		userA, userB, userB, userA,
	)
	if err != nil {
		return nil, fmt.Errorf("getting messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var creationDate string
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Body, &msg.Status, &creationDate); err != nil {
			return nil, err
		}
		msg.CreationDate, _ = time.Parse(time.RFC3339, creationDate)
		messages = append(messages, msg)
	}
	return messages, nil
}

func CreateMessage(db *sql.DB, id, senderID, receiverID, body string) (*Message, error) {
	now := time.Now().UTC()
	_, err := db.Exec(`
        INSERT INTO messages (id, sender_id, receiver_id, body, status, creation_date)
        VALUES (?, ?, ?, ?, 'sent', ?)`,
		id, senderID, receiverID, body, now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("creating message: %w", err)
	}
	return &Message{
		ID: id, SenderID: senderID, ReceiverID: receiverID,
		Body: body, Status: "sent", CreationDate: now,
	}, nil
}

func UpdateMessageStatus(db *sql.DB, messageID, status string) error {
	_, err := db.Exec(`
        UPDATE messages SET status = ? WHERE id = ?`, status, messageID)
	return err
}

func GetMessageByID(db *sql.DB, id string) (*Message, error) {
	row := db.QueryRow(`
        SELECT id, sender_id, receiver_id, body, status, creation_date
        FROM messages WHERE id = ?`, id)
	var msg Message
	var creationDate string
	err := row.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID,
		&msg.Body, &msg.Status, &creationDate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	msg.CreationDate, _ = time.Parse(time.RFC3339, creationDate)
	return &msg, nil
}

func GetUndeliveredMessages(db *sql.DB, receiverID string) ([]Message, error) {
	rows, err := db.Query(`
        SELECT id, sender_id, receiver_id, body, status, creation_date
        FROM messages
        WHERE receiver_id = ? AND status = 'sent'`, receiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		var creationDate string
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Body, &msg.Status, &creationDate); err != nil {
			return nil, err
		}
		msg.CreationDate, _ = time.Parse(time.RFC3339, creationDate)
		messages = append(messages, msg)
	}
	return messages, nil
}
