package assert

import (
	"fmt"
	"log"
	"runtime/debug"
)

// Mode defines the behavior of the assertion library
// In strict mode (default), assertions panic.
var StrictMode = true

// SuppressLogs disables assertion logging (useful for tests)
var SuppressLogs = false

// Check verifies a condition and returns an error if false.
// In strict mode (default), also panics with stack trace (fail-fast behavior per NASA Rule 5).
// Use for preconditions, postconditions, and invariants throughout hot paths.
// Returns nil if condition is true.
func Check(condition bool, msg string, args ...interface{}) error {
	if condition {
		return nil
	}

	formattedMsg := fmt.Sprintf(msg, args...)
	err := fmt.Errorf("ASSERTION FAILED: %s", formattedMsg)

	if !SuppressLogs {
		log.Printf("[CRITICAL] %v\nStack: %s", err, debug.Stack())
	}

	if StrictMode {
		panic(err)
	}

	return err
}

// NotNil checks that a pointer or interface is not nil.
// Returns an error (or panics in strict mode) if obj is nil.
func NotNil(obj interface{}, name string) error {
	return Check(obj != nil, "%s must not be nil", name)
}

// InRange checks that val is within [min, max] inclusive.
// Returns an error (or panics in strict mode) if val is out of range.
func InRange(val, min, max int, name string) error {
	return Check(val >= min && val <= max, "%s (%d) out of range [%d, %d]", name, val, min, max)
}

// True is an alias for Check for improved readability in assertion chains.
func True(condition bool, msg string, args ...interface{}) error {
	return Check(condition, msg, args...)
}
