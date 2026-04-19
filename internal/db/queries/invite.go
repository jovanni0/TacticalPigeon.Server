package queries

import (
	"database/sql"
	"fmt"
	"time"
)

type Invite struct {
	Token          string
	CreatedBy      string
	UsedBy         *string // pointer because it can be NULL
	CreationDate   time.Time
	ExpirationTime time.Time
}

func CreateInvite(db *sql.DB, token, createdBy string, expiresAt time.Time) error {
	_, err := db.Exec(`
        INSERT INTO invites (token, created_by, creation_date, expiration_time)
        VALUES (?, ?, ?, ?)`,
		token,
		createdBy,
		time.Now().UTC().Format(time.RFC3339),
		expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("creating invite: %w", err)
	}
	return nil
}

func GetInvite(db *sql.DB, token string) (*Invite, error) {
	row := db.QueryRow(`
        SELECT token, created_by, used_by, creation_date, expiration_time
        FROM invites
        WHERE token = ?`, token)

	var inv Invite
	var usedBy sql.NullString
	var creationDate, expirationTime string

	err := row.Scan(
		&inv.Token,
		&inv.CreatedBy,
		&usedBy,
		&creationDate,
		&expirationTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil // not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("getting invite: %w", err)
	}

	if usedBy.Valid {
		inv.UsedBy = &usedBy.String
	}
	inv.CreationDate, _ = time.Parse(time.RFC3339, creationDate)
	inv.ExpirationTime, _ = time.Parse(time.RFC3339, expirationTime)

	return &inv, nil
}

func MarkInviteUsed(db *sql.DB, token, userID string) error {
	_, err := db.Exec(`
        UPDATE invites SET used_by = ? WHERE token = ?`,
		userID, token,
	)
	return err
}
