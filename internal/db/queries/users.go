package queries

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID            string `json:"id"`
	DisplayName   string `json:"display_name"`
	ContactPolicy string `json:"contact_policy"`
	PublicKey     string `json:"public_key"`
	Role          string `json:"role"`
}

func CreateUser(db *sql.DB, inviteToken, userID, displayName, contactPolicy, publicKey string) error {
	// begin a transaction — both operations succeed or neither does
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	// if anything goes wrong, roll back all changes
	defer tx.Rollback()

	_, err = tx.Exec(`
        INSERT INTO users (id, display_name, contact_policy, public_key, role)
        VALUES (?, ?, ?, ?, 'user')`,
		userID, displayName, contactPolicy, publicKey,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	_, err = tx.Exec(`
        UPDATE invites SET used_by = ? WHERE token = ?`,
		userID, inviteToken,
	)
	if err != nil {
		return fmt.Errorf("marking invite used: %w", err)
	}

	// commit — only now are the changes written permanently
	return tx.Commit()
}

func GetUserByID(db *sql.DB, id string) (*User, error) {
	row := db.QueryRow(`
        SELECT id, display_name, contact_policy, public_key, role
        FROM users WHERE id = ?`, id)

	var u User
	err := row.Scan(&u.ID, &u.DisplayName, &u.ContactPolicy, &u.PublicKey, &u.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return &u, nil
}

func GetUserByPublicKey(db *sql.DB, publicKey string) (*User, error) {
	row := db.QueryRow(`
        SELECT id, display_name, contact_policy, public_key, role
        FROM users WHERE public_key = ?`, publicKey)

	var u User
	err := row.Scan(&u.ID, &u.DisplayName, &u.ContactPolicy, &u.PublicKey, &u.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by public key: %w", err)
	}
	return &u, nil
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query(`
        SELECT id, display_name, contact_policy, role
        FROM users
        ORDER BY display_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.ContactPolicy, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
