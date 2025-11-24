package fraud

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type DeviceFingerprint struct {
	DB *sql.DB
}

func NewDeviceFingerprint(db *sql.DB) *DeviceFingerprint {
	return &DeviceFingerprint{DB: db}
}

// ExtractDeviceInfo extracts device information from HTTP request
func (d *DeviceFingerprint) ExtractDeviceInfo(r *http.Request) (deviceHash, ipAddress, userAgent string) {
	ipAddress = r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent = r.UserAgent()

	// Create device hash from IP + User-Agent + custom fingerprint header
	fingerprintData := ipAddress + userAgent + r.Header.Get("X-Device-Fingerprint")
	hash := sha256.Sum256([]byte(fingerprintData))
	deviceHash = fmt.Sprintf("%x", hash)

	return
}

// RecordDevice stores device-user mapping
func (d *DeviceFingerprint) RecordDevice(userID, deviceHash, ipAddress, userAgent string) error {
	_, err := d.DB.Exec(`
		INSERT INTO device_fingerprints (user_id, device_hash, ip_address, user_agent, last_seen)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, device_hash) DO UPDATE
		SET last_seen = $5, ip_address = $3
	`, userID, deviceHash, ipAddress, userAgent, time.Now())
	return err
}

// CheckDeviceRisk flags if same device used by >5 users
func (d *DeviceFingerprint) CheckDeviceRisk(deviceHash string) (bool, error) {
	var userCount int
	err := d.DB.QueryRow(`
		SELECT COUNT(DISTINCT user_id)
		FROM device_fingerprints
		WHERE device_hash = $1
	`, deviceHash).Scan(&userCount)

	if err != nil {
		return false, err
	}

	// Flag if same device used by more than 5 users
	if userCount > 5 {
		return true, fmt.Errorf("suspicious device: used by %d users", userCount)
	}

	return false, nil
}
