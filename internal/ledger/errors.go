package ledger

import "errors"

var (
	// ErrChainTampered indicates the SHA-256 hash linkage is broken
	ErrChainTampered = errors.New("cryptographic chain linkage tampered: hash mismatch")

	// ErrInvalidSignature indicates an Ed25519 signature verification failure
	ErrInvalidSignature = errors.New("cryptographic signature invalid: authenticity unverified")

	// ErrSequenceGap indicates an unexpected jump in SequenceIndex
	ErrSequenceGap = errors.New("ledger sequence gap detected: missing intervening events")

	// ErrHashMismatch indicates the current event's hash does not match the recalculated data
	ErrHashMismatch = errors.New("event hash corrupted: data does not match stored hash")

	// ErrNoEvents indicates the run container exists but contains no events
	ErrNoEvents = errors.New("audit trail is empty: no events found for run")
)
