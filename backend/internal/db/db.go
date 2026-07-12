package db

import (
	"challenge-to-you/backend/internal/obs"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var log = obs.Default().Component("db")

type DB struct {
	conn *sql.DB

	stmtGetProfile             *sql.Stmt
	stmtUpdateProfile          *sql.Stmt
	stmtCheckToken             *sql.Stmt
	stmtInsertToken            *sql.Stmt
	stmtCheckChallenge         *sql.Stmt
	stmtInsertChallenge        *sql.Stmt
	stmtCheckMission           *sql.Stmt
	stmtInsertMission          *sql.Stmt
	stmtSaveActiveMission      *sql.Stmt
	stmtRemoveActiveMission    *sql.Stmt
	stmtGetActiveMissions      *sql.Stmt
	stmtGetCompletedChallenges *sql.Stmt
	stmtGetCompletedMissions   *sql.Stmt
}

type PlayerProfile struct {
	ID                int
	Name              string
	Luck              float64
	Reputation        int
	UnlockedParadigms string
	XP                int
	Level             int
	Title             string
}

func (p *PlayerProfile) IsParadigmUnlocked(paradigm string) bool {
	parts := strings.Split(p.UnlockedParadigms, ",")
	for _, part := range parts {
		if strings.TrimSpace(strings.ToUpper(part)) == strings.ToUpper(paradigm) {
			return true
		}
	}
	return false
}

var LevelThresholds = []int{
	0, 100, 250, 500, 800, 1200, 1700, 2300, 3000, 3800,
	4700, 5700, 6800, 8000, 9300, 10700, 12200, 13800, 15500, 17300,
}

func ComputeLevel(xp int) int {
	level := 1
	for i, threshold := range LevelThresholds {
		if xp >= threshold {
			level = i + 1
		} else {
			break
		}
	}
	return level
}

func TitleForLevel(level int) string {
	switch {
	case level >= 15:
		return "Archon"
	case level >= 10:
		return "Veteran"
	case level >= 5:
		return "Operative"
	case level >= 2:
		return "Initiate"
	default:
		return "Newcomer"
	}
}

type CompletedChallenge struct {
	ChallengeID string
	CompletedAt time.Time
	XPEarned    int
}

type CompletedMission struct {
	MissionID   string
	CompletedAt time.Time
}

type ActiveMission struct {
	MissionID   string
	PlayerID    string
	SessionJSON string
	StartedAt   time.Time
}

type SaveChecksum struct {
	Version   int    `json:"version"`
	Checksum  string `json:"checksum"`
	CreatedAt string `json:"created_at"`
}

const SaveVersion = 1

func computeChecksum(data string) string {
	h := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", h)
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	conn.SetMaxOpenConns(4)
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(5 * time.Minute)

	pragmas := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA temp_store = MEMORY;`,
		`PRAGMA cache_size = -64000;`,
		`PRAGMA busy_timeout = 5000;`,
	}
	for _, p := range pragmas {
		_, _ = conn.Exec(p)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS player_profile (
			id INTEGER PRIMARY KEY,
			name TEXT,
			luck REAL DEFAULT 1.0,
			reputation INTEGER DEFAULT 0,
			unlocked_paradigms TEXT DEFAULT 'MAGITECH',
			xp INTEGER DEFAULT 0,
			level INTEGER DEFAULT 1,
			title TEXT DEFAULT 'Newcomer',
			checksum TEXT DEFAULT '',
			save_version INTEGER DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS extracted_tokens (
			token TEXT PRIMARY KEY,
			extracted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS completed_challenges (
			challenge_id TEXT PRIMARY KEY,
			completed_at DATETIME,
			xp_earned INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS completed_missions (
			mission_id TEXT PRIMARY KEY,
			completed_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS active_missions (
			mission_id TEXT PRIMARY KEY,
			player_id TEXT,
			session_json TEXT,
			started_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS save_checksums (
			id INTEGER PRIMARY KEY,
			version INTEGER DEFAULT 1,
			checksum TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`INSERT OR IGNORE INTO player_profile (id, name, luck, reputation, unlocked_paradigms, xp, level, title)
		VALUES (1, 'Intruder', 1.0, 0, 'MAGITECH', 0, 1, 'Newcomer');`,
	}

	for _, query := range queries {
		if _, err := conn.Exec(query); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to execute initialization query: %w", err)
		}
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_completed_challenges_at ON completed_challenges(completed_at);`,
		`CREATE INDEX IF NOT EXISTS idx_completed_missions_at ON completed_missions(completed_at);`,
		`CREATE INDEX IF NOT EXISTS idx_active_missions_player ON active_missions(player_id);`,
		`CREATE INDEX IF NOT EXISTS idx_extracted_tokens_at ON extracted_tokens(extracted_at);`,
	}
	for _, idx := range indexes {
		_, _ = conn.Exec(idx)
	}

	migrations := []string{
		`ALTER TABLE player_profile ADD COLUMN checksum TEXT DEFAULT '';`,
		`ALTER TABLE player_profile ADD COLUMN save_version INTEGER DEFAULT 1;`,
	}
	for _, m := range migrations {
		_, _ = conn.Exec(m)
	}

	d := &DB{conn: conn}
	if err := d.prepareStatements(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) prepareStatements() error {
	var err error

	d.stmtGetProfile, err = d.conn.Prepare("SELECT id, name, luck, reputation, unlocked_paradigms, xp, level, title FROM player_profile WHERE id = 1")
	if err != nil {
		return fmt.Errorf("prepare get profile: %w", err)
	}

	d.stmtUpdateProfile, err = d.conn.Prepare("UPDATE player_profile SET name = ?, luck = ?, reputation = ?, unlocked_paradigms = ?, xp = ?, level = ?, title = ? WHERE id = 1")
	if err != nil {
		return fmt.Errorf("prepare update profile: %w", err)
	}

	d.stmtCheckToken, err = d.conn.Prepare("SELECT EXISTS(SELECT 1 FROM extracted_tokens WHERE token = ?)")
	if err != nil {
		return fmt.Errorf("prepare check token: %w", err)
	}

	d.stmtInsertToken, err = d.conn.Prepare("INSERT INTO extracted_tokens (token, extracted_at) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert token: %w", err)
	}

	d.stmtCheckChallenge, err = d.conn.Prepare("SELECT EXISTS(SELECT 1 FROM completed_challenges WHERE challenge_id = ?)")
	if err != nil {
		return fmt.Errorf("prepare check challenge: %w", err)
	}

	d.stmtInsertChallenge, err = d.conn.Prepare("INSERT INTO completed_challenges (challenge_id, completed_at, xp_earned) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert challenge: %w", err)
	}

	d.stmtCheckMission, err = d.conn.Prepare("SELECT EXISTS(SELECT 1 FROM completed_missions WHERE mission_id = ?)")
	if err != nil {
		return fmt.Errorf("prepare check mission: %w", err)
	}

	d.stmtInsertMission, err = d.conn.Prepare("INSERT INTO completed_missions (mission_id, completed_at) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert mission: %w", err)
	}

	d.stmtSaveActiveMission, err = d.conn.Prepare(`INSERT OR REPLACE INTO active_missions (mission_id, player_id, session_json, started_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare save active mission: %w", err)
	}

	d.stmtRemoveActiveMission, err = d.conn.Prepare("DELETE FROM active_missions WHERE mission_id = ?")
	if err != nil {
		return fmt.Errorf("prepare remove active mission: %w", err)
	}

	d.stmtGetActiveMissions, err = d.conn.Prepare("SELECT mission_id, player_id, session_json, started_at FROM active_missions")
	if err != nil {
		return fmt.Errorf("prepare get active missions: %w", err)
	}

	d.stmtGetCompletedChallenges, err = d.conn.Prepare("SELECT challenge_id, completed_at, xp_earned FROM completed_challenges ORDER BY completed_at")
	if err != nil {
		return fmt.Errorf("prepare get completed challenges: %w", err)
	}

	d.stmtGetCompletedMissions, err = d.conn.Prepare("SELECT mission_id, completed_at FROM completed_missions ORDER BY completed_at")
	if err != nil {
		return fmt.Errorf("prepare get completed missions: %w", err)
	}

	return nil
}

func (d *DB) Close() error {
	d.closeStmts()
	return d.conn.Close()
}

func (d *DB) closeStmts() {
	stmts := []*sql.Stmt{
		d.stmtGetProfile, d.stmtUpdateProfile, d.stmtCheckToken, d.stmtInsertToken,
		d.stmtCheckChallenge, d.stmtInsertChallenge, d.stmtCheckMission, d.stmtInsertMission,
		d.stmtSaveActiveMission, d.stmtRemoveActiveMission, d.stmtGetActiveMissions,
		d.stmtGetCompletedChallenges, d.stmtGetCompletedMissions,
	}
	for _, stmt := range stmts {
		if stmt != nil {
			stmt.Close()
		}
	}
}

func (d *DB) beginTx() (*sql.Tx, error) {
	return d.conn.Begin()
}

func verifyIntegrity(data string, expectedChecksum string) bool {
	if expectedChecksum == "" {
		return true
	}
	return computeChecksum(data) == expectedChecksum
}

func (d *DB) GetOrCreateProfile() (*PlayerProfile, error) {
	var p PlayerProfile
	var checksum string
	var saveVersion int
	err := d.stmtGetProfile.QueryRow().Scan(&p.ID, &p.Name, &p.Luck, &p.Reputation, &p.UnlockedParadigms, &p.XP, &p.Level, &p.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to scan player profile: %w", err)
	}

	profileData := fmt.Sprintf("%d|%s|%.4f|%d|%s|%d|%d|%s", p.ID, p.Name, p.Luck, p.Reputation, p.UnlockedParadigms, p.XP, p.Level, p.Title)
	_ = d.conn.QueryRow("SELECT checksum, save_version FROM player_profile WHERE id = 1").Scan(&checksum, &saveVersion)

	if !verifyIntegrity(profileData, checksum) {
		log.Warn("profile checksum mismatch", "expected", checksum, "got", computeChecksum(profileData), "data", profileData)
	}

	return &p, nil
}

func (d *DB) SaveProfile(p *PlayerProfile) error {
	p.Level = ComputeLevel(p.XP)
	p.Title = TitleForLevel(p.Level)

	profileData := fmt.Sprintf("%d|%s|%.4f|%d|%s|%d|%d|%s", p.ID, p.Name, p.Luck, p.Reputation, p.UnlockedParadigms, p.XP, p.Level, p.Title)
	checksum := computeChecksum(profileData)

	tx, err := d.beginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtUpdateProfile).Exec(p.Name, p.Luck, p.Reputation, p.UnlockedParadigms, p.XP, p.Level, p.Title)
	if err != nil {
		return fmt.Errorf("failed to update player profile: %w", err)
	}

	_, err = tx.Exec("UPDATE player_profile SET checksum = ?, save_version = ? WHERE id = 1", checksum, SaveVersion)
	if err != nil {
		return fmt.Errorf("failed to update profile checksum: %w", err)
	}

	return tx.Commit()
}

func (d *DB) AddXP(amount int) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var id, level int
	var name, unlocked, title string
	var luck float64
	var rep, xp int
	err = tx.QueryRow("SELECT id, name, luck, reputation, unlocked_paradigms, xp, level, title FROM player_profile WHERE id = 1").Scan(&id, &name, &luck, &rep, &unlocked, &xp, &level, &title)
	if err != nil {
		return fmt.Errorf("failed to read profile: %w", err)
	}

	xp += amount
	level = ComputeLevel(xp)
	title = TitleForLevel(level)

	profileData := fmt.Sprintf("%d|%s|%.4f|%d|%s|%d|%d|%s", id, name, luck, rep, unlocked, xp, level, title)
	checksum := computeChecksum(profileData)

	_, err = tx.Exec("UPDATE player_profile SET luck = ?, reputation = ?, unlocked_paradigms = ?, xp = ?, level = ?, title = ?, checksum = ?, save_version = ? WHERE id = 1",
		luck, rep, unlocked, xp, level, title, checksum, SaveVersion)
	if err != nil {
		return fmt.Errorf("failed to add xp: %w", err)
	}

	return tx.Commit()
}

func (d *DB) RecordToken(token string) (bool, error) {
	var exists bool
	err := d.stmtCheckToken.QueryRow(token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	if exists {
		return false, nil
	}

	tx, err := d.beginTx()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtInsertToken).Exec(token, time.Now())
	if err != nil {
		return false, fmt.Errorf("failed to insert token: %w", err)
	}

	return true, tx.Commit()
}

func (d *DB) RecordChallengeCompletion(challengeID string, xpEarned int) (bool, error) {
	var exists bool
	err := d.stmtCheckChallenge.QueryRow(challengeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check challenge existence: %w", err)
	}

	if exists {
		return false, nil
	}

	tx, err := d.beginTx()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtInsertChallenge).Exec(challengeID, time.Now(), xpEarned)
	if err != nil {
		return false, fmt.Errorf("failed to record challenge completion: %w", err)
	}

	return true, tx.Commit()
}

func (d *DB) IsChallengeCompleted(challengeID string) bool {
	var exists bool
	_ = d.stmtCheckChallenge.QueryRow(challengeID).Scan(&exists)
	return exists
}

func (d *DB) GetCompletedChallenges() ([]CompletedChallenge, error) {
	rows, err := d.stmtGetCompletedChallenges.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query completed challenges: %w", err)
	}
	defer rows.Close()

	var results []CompletedChallenge
	for rows.Next() {
		var cc CompletedChallenge
		if err := rows.Scan(&cc.ChallengeID, &cc.CompletedAt, &cc.XPEarned); err != nil {
			return nil, fmt.Errorf("failed to scan completed challenge: %w", err)
		}
		results = append(results, cc)
	}
	return results, nil
}

func (d *DB) RecordMissionCompletion(missionID string) (bool, error) {
	var exists bool
	err := d.stmtCheckMission.QueryRow(missionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}

	if exists {
		return false, nil
	}

	tx, err := d.beginTx()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtInsertMission).Exec(missionID, time.Now())
	if err != nil {
		return false, fmt.Errorf("failed to record mission completion: %w", err)
	}

	return true, tx.Commit()
}

func (d *DB) IsMissionCompleted(missionID string) bool {
	var exists bool
	_ = d.stmtCheckMission.QueryRow(missionID).Scan(&exists)
	return exists
}

func (d *DB) GetCompletedMissions() ([]CompletedMission, error) {
	rows, err := d.stmtGetCompletedMissions.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query completed missions: %w", err)
	}
	defer rows.Close()

	var results []CompletedMission
	for rows.Next() {
		var cm CompletedMission
		if err := rows.Scan(&cm.MissionID, &cm.CompletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan completed mission: %w", err)
		}
		results = append(results, cm)
	}
	return results, nil
}

func (d *DB) SaveActiveMission(missionID, playerID string, session interface{}) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	tx, err := d.beginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtSaveActiveMission).Exec(missionID, playerID, string(data), time.Now())
	if err != nil {
		return fmt.Errorf("failed to save active mission: %w", err)
	}

	return tx.Commit()
}

func (d *DB) RemoveActiveMission(missionID string) error {
	tx, err := d.beginTx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Stmt(d.stmtRemoveActiveMission).Exec(missionID)
	if err != nil {
		return fmt.Errorf("failed to remove active mission: %w", err)
	}

	return tx.Commit()
}

func (d *DB) GetActiveMissions() ([]ActiveMission, error) {
	rows, err := d.stmtGetActiveMissions.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query active missions: %w", err)
	}
	defer rows.Close()

	var results []ActiveMission
	for rows.Next() {
		var am ActiveMission
		if err := rows.Scan(&am.MissionID, &am.PlayerID, &am.SessionJSON, &am.StartedAt); err != nil {
			return nil, fmt.Errorf("failed to scan active mission: %w", err)
		}
		results = append(results, am)
	}
	return results, nil
}

func (d *DB) VerifySave() (bool, error) {
	var p PlayerProfile
	var checksum string
	var saveVersion int
	err := d.stmtGetProfile.QueryRow().Scan(&p.ID, &p.Name, &p.Luck, &p.Reputation, &p.UnlockedParadigms, &p.XP, &p.Level, &p.Title)
	if err != nil {
		return false, fmt.Errorf("failed to query profile: %w", err)
	}

	row := d.conn.QueryRow("SELECT checksum, save_version FROM player_profile WHERE id = 1")
	if err := row.Scan(&checksum, &saveVersion); err != nil {
		return false, nil
	}

	if checksum == "" {
		return true, nil
	}

	profileData := fmt.Sprintf("%d|%s|%.4f|%d|%s|%d|%d|%s", p.ID, p.Name, p.Luck, p.Reputation, p.UnlockedParadigms, p.XP, p.Level, p.Title)
	return computeChecksum(profileData) == checksum, nil
}

func (d *DB) CheckSaveVersion() (int, error) {
	var version int
	err := d.conn.QueryRow("SELECT COALESCE(save_version, 0) FROM player_profile WHERE id = 1").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
