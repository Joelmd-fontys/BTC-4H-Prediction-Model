package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return database, nil
}
