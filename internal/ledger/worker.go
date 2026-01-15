package ledger

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/vouch/internal/assert"
	"github.com/yourname/vouch/internal/crypto"
	"github.com/yourname/vouch/internal/proxy"
)

// Worker processes events asynchronously without blocking the proxy
type Worker struct {
	eventChannel chan proxy.Event
	db           *DB
	signer       *crypto.Signer
	runID        string
	taskStates   map[string]string // Track task state changes
	isUnhealthy  atomic.Bool       // Health sentinel
}

// NewWorker creates a new async ledger worker with a buffered channel
func NewWorker(bufferSize int, dbPath, keyPath string) (*Worker, error) {
	// Initialize database
	db, err := NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	// Initialize signer (loads or generates keypair)
	signer, err := crypto.NewSigner(keyPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing signer: %w", err)
	}

	return &Worker{
		eventChannel: make(chan proxy.Event, bufferSize),
		db:           db,
		signer:       signer,
		taskStates:   make(map[string]string),
	}, nil
}

// GetDB returns the database instance
func (w *Worker) GetDB() *DB {
	return w.db
}

// IsHealthy returns the current health status of the worker
func (w *Worker) IsHealthy() bool {
	return !w.isUnhealthy.Load()
}

// GetSigner returns the signer instance
func (w *Worker) GetSigner() *crypto.Signer {
	return w.signer
}

// Start initializes the worker, loads existing runs or creates a genesis block, and starts event processing.
func (w *Worker) Start() error {
	// Check for existing runs
	hasRuns, err := w.db.HasRuns()
	if err != nil {
		return fmt.Errorf("checking for existing runs: %w", err)
	}

	if !hasRuns {
		// Create genesis block
		runID, err := CreateGenesisBlock(w.db, w.signer, "Vouch-Agent")
		if err != nil {
			return fmt.Errorf("creating genesis block: %w", err)
		}
		w.runID = runID
		log.Printf("Genesis block created (Run ID: %s)", runID[:8])
		log.Printf("Public key: %s", w.signer.GetPublicKey()[:32]+"...")
	} else {
		// Load existing run ID
		runID, err := w.db.GetRunID()
		if err != nil {
			return fmt.Errorf("loading run ID: %w", err)
		}
		w.runID = runID
		log.Printf("Loaded existing run (Run ID: %s)", runID[:8])
	}

	// Start async worker
	go w.processEvents()
	log.Println("Async worker started with database persistence")

	return nil
}

// Submit sends an event to the worker for processing (non-blocking)
func (w *Worker) Submit(event proxy.Event) {
	// Backpressure Awareness: Log a high-visibility warning if the buffer is > 80% full
	capacity := cap(w.eventChannel)
	current := len(w.eventChannel)
	if capacity > 0 && float64(current)/float64(capacity) >= 0.8 {
		log.Printf("========================================================")
		log.Printf("[BACKPRESSURE] Ledger buffer at %d/%d (>=80%%) capacity", current, capacity)
		log.Printf("[BACKPRESSURE] Throttling agent requests to prevent loss")
		log.Printf("========================================================")
	}

	w.eventChannel <- event
}

// Close shuts down the worker and closes the database
func (w *Worker) Close() error {
	close(w.eventChannel)
	return w.db.Close()
}

// processEvents is the main worker loop
func (w *Worker) processEvents() {
	for event := range w.eventChannel {
		if err := w.persistEvent(&event); err != nil {
			log.Printf("Error persisting event: %v", err)
			continue
		}

		// Log to console for visibility
		timestamp := event.Timestamp.Format("15:04:05.000")

		if event.WasBlocked {
			log.Printf("[%s] BLOCKED | %s | Seq: %d | Hash: %s",
				timestamp, event.Method, event.SeqIndex, event.CurrentHash[:16])
		} else if event.EventType == "tool_call" {
			log.Printf("[%s] CALL    | %s | Seq: %d | Hash: %s",
				timestamp, event.Method, event.SeqIndex, event.CurrentHash[:16])
			if event.TaskID != "" {
				// Check for task state change
				oldState, exists := w.taskStates[event.TaskID]
				if exists && oldState != event.TaskState {
					log.Printf("  Task %s: %s -> %s", event.TaskID, oldState, event.TaskState)

					// If task completed, create a task_completed event
					if event.TaskState == "completed" || event.TaskState == "failed" || event.TaskState == "cancelled" {
						w.createTaskCompletionEvent(event.TaskID, event.TaskState)

						// Terminal State Cleanup: Prevent memory leak
						delete(w.taskStates, event.TaskID)
						log.Printf(" [CLEANUP] Task %s state purged from memory", event.TaskID)
						continue // State already purged, don't re-add below
					}
				}
				w.taskStates[event.TaskID] = event.TaskState
			}
		} else if event.EventType == "tool_response" {
			log.Printf("[%s] RESPONSE| %s | Seq: %d | Hash: %s",
				timestamp, event.Method, event.SeqIndex, event.CurrentHash[:16])
			if event.TaskID != "" {
				// Check for task state change
				oldState, exists := w.taskStates[event.TaskID]
				if exists && oldState != event.TaskState {
					log.Printf("  Task %s: %s -> %s", event.TaskID, oldState, event.TaskState)

					// If task completed, create a task_completed event
					if event.TaskState == "completed" || event.TaskState == "failed" || event.TaskState == "cancelled" {
						w.createTaskCompletionEvent(event.TaskID, event.TaskState)

						// Terminal State Cleanup: Prevent memory leak
						delete(w.taskStates, event.TaskID)
						log.Printf(" [CLEANUP] Task %s state purged from memory", event.TaskID)
						continue // State already purged, don't re-add below
					}
				}
				w.taskStates[event.TaskID] = event.TaskState
			}
		}
	}
}

// persistEvent prepares, hashes, signs and stores an event in the database
func (w *Worker) persistEvent(event *proxy.Event) error {
	// 1. Assign sequence index
	runStats, err := w.db.GetRunStats(w.runID)
	if err != nil {
		return fmt.Errorf("getting run stats: %w", err)
	}
	event.SeqIndex = runStats.TotalEvents
	event.RunID = w.runID

	// 2. Get previous hash
	var prevHash string
	if event.SeqIndex == 0 {
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	} else {
		// GetLastEvent returns (seqIndex, currentHash, err)
		_, lastHash, err := w.db.GetLastEvent(w.runID)
		if err != nil {
			return fmt.Errorf("getting last event: %w", err)
		}

		// Safety Assertion: Verify prev_hash is non-empty for non-genesis blocks
		if err := assert.Check(lastHash != "", "prev_hash must be non-empty for non-genesis blocks", "seq", event.SeqIndex); err != nil {
			return err
		}

		prevHash = lastHash
	}
	event.PrevHash = prevHash

	if err := assert.Check(len(event.PrevHash) == 64, "invalid prev_hash length", "len", len(event.PrevHash)); err != nil {
		return err
	}

	// 3. Normalize timestamp for consistent hashing
	// Use RFC3339Nano for maximum precision during serialization
	tsStr := event.Timestamp.Format(time.RFC3339Nano)

	// 4. Calculate hash using normalized payload and JCS
	payload := map[string]interface{}{
		"id":         event.ID,
		"run_id":     event.RunID,
		"seq_index":  event.SeqIndex,
		"timestamp":  tsStr,
		"actor":      event.Actor,
		"event_type": event.EventType,
		"method":     event.Method,
		"params":     event.Params,
		"response":   event.Response,
		"task_id":    event.TaskID,
		"task_state": event.TaskState,
		"parent_id":  event.ParentID,
		"policy_id":  event.PolicyID,
		"risk_level": event.RiskLevel,
	}

	currentHash, err := crypto.CalculateEventHash(event.PrevHash, payload)
	if err != nil {
		return fmt.Errorf("calculating hash: %w", err)
	}
	event.CurrentHash = currentHash

	// 5. Sign the hash
	signature, err := w.signer.SignHash(currentHash)
	if err != nil {
		return fmt.Errorf("signing hash: %w", err)
	}
	event.Signature = signature

	// Insert into database
	if err := insertEvent(w.db, *event); err != nil {
		// Health Sentinel: Mark system as unhealthy on database failure
		w.isUnhealthy.Store(true)
		log.Printf("========================================================")
		log.Printf("[CRITICAL] DATABASE WRITE FAILURE: %v", err)
		log.Printf("[CRITICAL] Setting system health to UNHEALTHY")
		log.Printf("========================================================")
		return err
	}

	return nil
}

// createTaskCompletionEvent creates a task_completed event when a task finishes
func (w *Worker) createTaskCompletionEvent(taskID string, state string) {
	event := proxy.Event{
		ID:        uuid.New().String()[:8],
		Timestamp: time.Now(),
		EventType: "task_terminal",
		Method:    "vouch:task_state",
		Params: map[string]interface{}{
			"task_id": taskID,
			"state":   state,
		},
		TaskID:    taskID,
		TaskState: state,
	}

	// Direct call to persist to avoid channel recursion if needed,
	// but channel is fine for this small event.
	w.Submit(event)
}
