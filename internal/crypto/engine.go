package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/ucarion/jcs"
)

// CalculateEventHash ensures deterministic hashing across any model/platform
// Uses RFC 8785 (JSON Canonicalization Scheme) for consistent serialization
func CalculateEventHash(prevHash string, payload interface{}) (string, error) {
	// 1. First marshal to JSON to normalize the data structure
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// 2. Unmarshal to a clean interface{} for JCS
	var normalized interface{}
	if err := json.Unmarshal(jsonBytes, &normalized); err != nil {
		return "", err
	}

	// 3. Canonicalize using JCS (RFC 8785)
	// This ensures identical output regardless of key order
	canonicalJSON, err := jcs.Format(normalized)
	if err != nil {
		return "", err
	}

	// 4. Hash(Prev + Current)
	hasher := sha256.New()
	hasher.Write([]byte(prevHash))
	hasher.Write([]byte(canonicalJSON))

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
