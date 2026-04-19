package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// SQLite allows only one writer at a time by default.
	// This prevents "database is locked" errors under concurrent requests.
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id             TEXT PRIMARY KEY,
        display_name   TEXT NOT NULL,
        contact_policy TEXT NOT NULL DEFAULT 'open',
        public_key     TEXT NOT NULL,
        role           TEXT NOT NULL DEFAULT 'user'
    );

    CREATE TABLE IF NOT EXISTS invites (
        token           TEXT PRIMARY KEY,
        created_by      TEXT NOT NULL REFERENCES users(id),
        used_by         TEXT REFERENCES users(id),
        creation_date   TEXT NOT NULL,
        expiration_time TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS sessions (
        token      TEXT PRIMARY KEY,
        user_id    TEXT NOT NULL REFERENCES users(id),
        created_at TEXT NOT NULL,
        expires_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS auth_challenges (
        token      TEXT PRIMARY KEY,
        user_id    TEXT NOT NULL REFERENCES users(id),
        expires_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS messages (
        id            TEXT PRIMARY KEY,
        sender_id     TEXT NOT NULL REFERENCES users(id),
        receiver_id   TEXT NOT NULL REFERENCES users(id),
        body          TEXT NOT NULL,
        status        TEXT NOT NULL DEFAULT 'sent',
        creation_date TEXT NOT NULL
    );
    `

	_, err := db.Exec(schema)
	return err
}
