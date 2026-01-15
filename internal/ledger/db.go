package ledger

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection and initializes the schema
func NewDB(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Execute schema
	schemaPath := "schema.sql"
	if !filepath.IsAbs(schemaPath) {
		wd, err := os.Getwd()
		if err == nil {
			schemaPath = filepath.Join(wd, schemaPath)
		}
	}

	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	if _, err := conn.Exec(string(schemaSQL)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("executing schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// InsertRun creates a new run record
func (db *DB) InsertRun(id, agentName, genesisHash, ledgerPubKey string) error {
	query := `INSERT INTO runs (id, agent_name, genesis_hash, ledger_pub_key) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, id, agentName, genesisHash, ledgerPubKey)
	if err != nil {
		return fmt.Errorf("inserting run: %w", err)
	}
	return nil
}

// InsertEvent inserts a new event into the ledger
func (db *DB) InsertEvent(
	id, runID string,
	seqIndex int,
	timestamp, actor, eventType, method, params, response, taskID, taskState, prevHash, currentHash, signature string,
) error {
	query := `
		INSERT INTO events (
			id, run_id, seq_index, timestamp, actor, event_type, method, params, response,
			task_id, task_state, prev_hash, current_hash, signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		id, runID, seqIndex, timestamp, actor, eventType, method, params, response,
		taskID, taskState, prevHash, currentHash, signature,
	)
	if err != nil {
		return fmt.Errorf("inserting event: %w", err)
	}
	return nil
}

// GetLastEvent retrieves the most recent event for a given run
func (db *DB) GetLastEvent(runID string) (seqIndex int, currentHash string, err error) {
	query := `SELECT seq_index, current_hash FROM events WHERE run_id = ? ORDER BY seq_index DESC LIMIT 1`
	err = db.conn.QueryRow(query, runID).Scan(&seqIndex, &currentHash)
	if err == sql.ErrNoRows {
		// No events yet, return defaults
		return -1, "", nil
	}
	if err != nil {
		return 0, "", fmt.Errorf("querying last event: %w", err)
	}
	return seqIndex, currentHash, nil
}

// HasRuns checks if any runs exist in the database
func (db *DB) HasRuns() (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking runs: %w", err)
	}
	return count > 0, nil
}

// GetRunID retrieves the most recent run ID
func (db *DB) GetRunID() (string, error) {
	var runID string
	err := db.conn.QueryRow("SELECT id FROM runs ORDER BY started_at DESC LIMIT 1").Scan(&runID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("querying run ID: %w", err)
	}
	return runID, nil
}
