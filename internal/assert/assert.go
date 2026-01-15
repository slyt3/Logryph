//go:build !release
// +build !release

package assert

import (
	"errors"
	"fmt"
	"log"
	"runtime"
)

// Check verifies a safety-critical condition.
// If the condition is false, it logs the failure context (caller info, message, fields)
// and returns a descriptive error. It never panics.
func Check(condition bool, msg string, fields ...interface{}) error {
	if condition {
		return nil
	}

	// Capture caller info (filename, line)
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format context fields for logging
	context := ""
	if len(fields) > 0 {
		context = fmt.Sprintf(" | context: %v", fields)
	}

	errMsg := fmt.Sprintf("[ASSERTION FAILURE] %s:%d: %s%s", file, line, msg, context)
	log.Println(errMsg)

	return errors.New(errMsg)
}
