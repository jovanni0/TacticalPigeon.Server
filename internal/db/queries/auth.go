package queries

import (
	"database/sql"
	"fmt"
	"time"
)

func CreateChallenge(db *sql.DB, challengeToken, userID string, expiresAt time.Time) error {
	_, err := db.Exec(`
        INSERT INTO auth_challenges (token, user_id, expires_at)
        VALUES (?, ?, ?)`,
		challengeToken,
		userID,
		expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func GetChallenge(db *sql.DB, token string) (userID string, expiresAt time.Time, err error) {
	row := db.QueryRow(`
        SELECT user_id, expires_at FROM auth_challenges
        WHERE token = ?`, token)

	var expiresAtStr string
	err = row.Scan(&userID, &expiresAtStr)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("getting challenge: %w", err)
	}

	expiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	return
}

func DeleteChallenge(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM auth_challenges WHERE token = ?`, token)
	return err
}

func CreateSession(db *sql.DB, sessionToken, userID string, expiresAt time.Time) error {
	_, err := db.Exec(`
        INSERT INTO sessions (token, user_id, created_at, expires_at)
        VALUES (?, ?, ?, ?)`,
		sessionToken,
		userID,
		time.Now().UTC().Format(time.RFC3339),
		expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

func GetSession(db *sql.DB, token string) (userID string, expiresAt time.Time, err error) {
	row := db.QueryRow(`
        SELECT user_id, expires_at FROM sessions
        WHERE token = ?`, token)

	var expiresAtStr string
	err = row.Scan(&userID, &expiresAtStr)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("getting session: %w", err)
	}

	expiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
	return
}
