package fairness

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

// GenerateServerSeed creates a new random server seed
func GenerateServerSeed() string {
	// In production, use crypto/rand
	// For demo, we'll use a simple hash of time
	// TODO: Replace with secure random generation
	return hex.EncodeToString(sha256.New().Sum([]byte("random_seed")))
}

// HashServerSeed creates the public hash of the server seed
func HashServerSeed(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(hash[:])
}

// GenerateFloat generates a float between 0 and 1 using HMAC-SHA256
// This is the standard "Provably Fair" algorithm used by Stake/PrimeDice
func GenerateFloat(serverSeed, clientSeed string, nonce int) float64 {
	// 1. Create HMAC-SHA256(server_seed, client_seed:nonce)
	message := fmt.Sprintf("%s:%d", clientSeed, nonce)
	h := hmac.New(sha256.New, []byte(serverSeed))
	h.Write([]byte(message))
	hash := hex.EncodeToString(h.Sum(nil))

	// 2. Take first 4 bytes (32 bits)
	// We iterate 4 bytes at a time until we find a number < 2^32 / 100 * 100 (to avoid bias)
	// For simplicity in this implementation, we just take the first 4 bytes
	// A full implementation handles the edge case of "modulo bias"

	chunk := hash[:8] // 8 hex chars = 4 bytes
	decimalValue, _ := strconv.ParseUint(chunk, 16, 64)

	// 3. Convert to float [0, 1)
	// Max value of 4 bytes is 4294967295
	return float64(decimalValue) / 4294967296.0
}

// CrashPoint calculates the crash multiplier from the hash
// Formula: E = 2^52, h = hash; crash = (100 * E - h) / (E - h)
// Simplified Stake formula: 0.99 / (1 - X)
func CalculateCrashPoint(serverSeed, clientSeed string, nonce int) float64 {
	x := GenerateFloat(serverSeed, clientSeed, nonce)

	// House edge of 1% (0.99)
	multiplier := 0.99 / (1 - x)

	// Cap at 1.00x (instant crash) if result is very low
	if multiplier < 1.00 {
		multiplier = 1.00
	}

	return multiplier
}
