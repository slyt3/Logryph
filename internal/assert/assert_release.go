//go:build release
// +build release

package assert

// Check is a zero-cost nop in release builds.
func Check(condition bool, msg string, fields ...interface{}) error {
	return nil
}
