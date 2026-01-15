package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

// Signer handles Ed25519 signing operations
type Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewSigner creates a new signer, loading or generating keys as needed
func NewSigner(keyPath string) (*Signer, error) {
	// Try to load existing key
	privateKey, err := loadPrivateKey(keyPath)
	if err != nil {
		// Generate new keypair
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating keypair: %w", err)
		}

		// Save private key
		if err := savePrivateKey(keyPath, privateKey); err != nil {
			return nil, fmt.Errorf("saving private key: %w", err)
		}

		return &Signer{
			privateKey: privateKey,
			publicKey:  publicKey,
		}, nil
	}

	// Derive public key from private key
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Signer{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// SignHash signs a hash string and returns the signature as hex
func (s *Signer) SignHash(hash string) (string, error) {
	hashBytes := []byte(hash)
	signature := ed25519.Sign(s.privateKey, hashBytes)
	return hex.EncodeToString(signature), nil
}

// GetPublicKey returns the public key as hex string
func (s *Signer) GetPublicKey() string {
	return hex.EncodeToString(s.publicKey)
}

// VerifySignature verifies a signature against a hash
func (s *Signer) VerifySignature(hash, signatureHex string) bool {
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}
	return ed25519.Verify(s.publicKey, []byte(hash), signature)
}

// loadPrivateKey loads a private key from file (hex-encoded)
func loadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keyBytes, err := hex.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("decoding key: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", ed25519.PrivateKeySize, len(keyBytes))
	}

	return ed25519.PrivateKey(keyBytes), nil
}

// savePrivateKey saves a private key to file (hex-encoded)
func savePrivateKey(path string, key ed25519.PrivateKey) error {
	hexKey := hex.EncodeToString(key)
	return os.WriteFile(path, []byte(hexKey), 0600) // Restrictive permissions
}
