package core

import (
	"sync"

	"github.com/slyt3/Vouch/internal/ledger"
	"github.com/slyt3/Vouch/internal/observer"
	"github.com/slyt3/Vouch/internal/proxy"
)

// Engine is the central state manager for Vouch
// Engine is the central state manager for Vouch
type Engine struct {
	Worker          *ledger.Worker
	ActiveTasks     *sync.Map           // task_id -> state
	Policy          *proxy.PolicyConfig // Deprecated: Only for temporary compatibility
	Observer        *observer.ObserverEngine
	StallSignals    *sync.Map // Maps event ID to approval channel
	LastEventByTask *sync.Map // task_id -> last_event_id
}

// NewEngine creates a new core state engine
func NewEngine(worker *ledger.Worker, policy *proxy.PolicyConfig, obs *observer.ObserverEngine) *Engine {
	return &Engine{
		Worker:          worker,
		Policy:          policy,
		Observer:        obs,
		ActiveTasks:     &sync.Map{},
		StallSignals:    &sync.Map{},
		LastEventByTask: &sync.Map{},
	}
}
