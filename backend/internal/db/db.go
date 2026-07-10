package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var databaseConn *sql.DB

type PlayerProfile struct {
	ID                int
	Name              string
	Luck              float64
	Reputation        int
	UnlockedParadigms string // Comma-separated paradigms, e.g. "MAGITECH,CYBERPUNK"
}

// IsParadigmUnlocked returns true if the specified paradigm exists in the unlocked string
func (p *PlayerProfile) IsParadigmUnlocked(paradigm string) bool {
	parts := strings.Split(p.UnlockedParadigms, ",")
	for _, part := range parts {
		if strings.TrimSpace(strings.ToUpper(part)) == strings.ToUpper(paradigm) {
			return true
		}
	}
	return false
}

// InitDB initializes SQLite and sets up the schemas
func InitDB(dbPath string) error {
	var err error
	databaseConn, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS player_profile (
			id INTEGER PRIMARY KEY,
			name TEXT,
			luck REAL DEFAULT 1.0,
			reputation INTEGER DEFAULT 0,
			unlocked_paradigms TEXT DEFAULT 'MAGITECH'
		);`,
		`CREATE TABLE IF NOT EXISTS extracted_tokens (
			token TEXT PRIMARY KEY,
			extracted_at DATETIME
		);`,
		`INSERT OR IGNORE INTO player_profile (id, name, luck, reputation, unlocked_paradigms)
		VALUES (1, 'Intruder', 1.0, 0, 'MAGITECH');`,
	}

	for _, query := range queries {
		if _, err := databaseConn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute initialization query: %w", err)
		}
	}

	return nil
}

// CloseDB closes the SQLite database connection
func CloseDB() error {
	if databaseConn != nil {
		return databaseConn.Close()
	}
	return nil
}

// GetOrCreateProfile retrieves the single player profile from database
func GetOrCreateProfile() (*PlayerProfile, error) {
	if databaseConn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	row := databaseConn.QueryRow("SELECT id, name, luck, reputation, unlocked_paradigms FROM player_profile WHERE id = 1")
	var p PlayerProfile
	err := row.Scan(&p.ID, &p.Name, &p.Luck, &p.Reputation, &p.UnlockedParadigms)
	if err != nil {
		return nil, fmt.Errorf("failed to scan player profile: %w", err)
	}

	return &p, nil
}

// SaveProfile saves mutations to the player profile
func SaveProfile(p *PlayerProfile) error {
	if databaseConn == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := databaseConn.Exec(
		"UPDATE player_profile SET name = ?, luck = ?, reputation = ?, unlocked_paradigms = ? WHERE id = 1",
		p.Name, p.Luck, p.Reputation, p.UnlockedParadigms,
	)
	if err != nil {
		return fmt.Errorf("failed to update player profile: %w", err)
	}
	return nil
}

// RecordToken inserts a new extracted Logos Token. Returns true if newly inserted, false if already extracted.
func RecordToken(token string) (bool, error) {
	if databaseConn == nil {
		return false, fmt.Errorf("database not initialized")
	}

	var exists bool
	err := databaseConn.QueryRow("SELECT EXISTS(SELECT 1 FROM extracted_tokens WHERE token = ?)", token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	if exists {
		return false, nil
	}

	_, err = databaseConn.Exec("INSERT INTO extracted_tokens (token, extracted_at) VALUES (?, ?)", token, time.Now())
	if err != nil {
		return false, fmt.Errorf("failed to insert token: %w", err)
	}

	return true, nil
}
