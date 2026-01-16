package ledger

import "github.com/slyt3/Vouch/internal/models"

// EventRepository defines the storage interface for the Vouch ledger.
// This allows swapping SQLite for Postgres/dqlite in the future without changing core logic.
type EventRepository interface {
	// Writer
	StoreEvent(event *models.Event) error
	InsertRun(id, agent, genesisHash, pubKey string) error

	// Reader
	GetLastEvent(runID string) (uint64, string, error)
	GetEventByID(eventID string) (*models.Event, error)
	GetAllEvents(runID string) ([]models.Event, error)
	GetRecentEvents(runID string, limit int) ([]models.Event, error)
	GetEventsByTaskID(taskID string) ([]models.Event, error)
	GetRiskEvents() ([]models.Event, error)

	// Meta
	HasRuns() (bool, error)
	GetRunID() (string, error)
	GetRunInfo(runID string) (agent, genesisHash, pubKey string, err error)

	// Stats
	GetRunStats(runID string) (*RunStats, error)
	GetGlobalStats() (*GlobalStats, error)

	// Lifecycle
	Close() error
}
