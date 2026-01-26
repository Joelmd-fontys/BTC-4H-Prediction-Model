package store

import (
	"database/sql"
	"fmt"
)

func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path) // for modernc.org/sqlite
	if err != nil {
		return nil, err
	}

	// Basic sanity check
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
