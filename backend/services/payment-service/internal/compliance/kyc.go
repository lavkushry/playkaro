package compliance

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// KYC Status
const (
	KYCStatusPending  = "PENDING"
	KYCStatusVerified = "VERIFIED"
	KYCStatusRejected = "REJECTED"
)

// Document Types
const (
	DocTypeAadhaar = "AADHAAR"
	DocTypePAN     = "PAN"
)

// KYCRequest represents a user's KYC submission
type KYCRequest struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	DocumentType   string    `json:"document_type" db:"document_type"`
	DocumentNumber string    `json:"document_number" db:"document_number"` // Encrypted
	ImageURL       string    `json:"image_url" db:"image_url"`             // Encrypted
	Status         string    `json:"status" db:"status"`
	AdminNotes     string    `json:"admin_notes" db:"admin_notes"`
	VerifiedAt     *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type KYCService struct {
	DB *sql.DB
}

func NewKYCService(db *sql.DB) *KYCService {
	return &KYCService{DB: db}
}

// SubmitKYC handles document submission
func (s *KYCService) SubmitKYC(userID, docType, docNumber, imageURL string) (*KYCRequest, error) {
	// Validate inputs
	if docType != DocTypeAadhaar && docType != DocTypePAN {
		return nil, errors.New("invalid document type")
	}

	// Encrypt sensitive data
	encDocNumber, err := EncryptPII(docNumber)
	if err != nil {
		return nil, err
	}
	encImageURL, err := EncryptPII(imageURL)
	if err != nil {
		return nil, err
	}

	// Check for existing pending request
	var count int
	err = s.DB.QueryRow("SELECT COUNT(*) FROM kyc_requests WHERE user_id = $1 AND status = $2", userID, KYCStatusPending).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("kyc request already pending")
	}

	req := &KYCRequest{
		ID:             uuid.New().String(),
		UserID:         userID,
		DocumentType:   docType,
		DocumentNumber: encDocNumber,
		ImageURL:       encImageURL,
		Status:         KYCStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = s.DB.Exec(`
		INSERT INTO kyc_requests (id, user_id, document_type, document_number, image_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, req.ID, req.UserID, req.DocumentType, req.DocumentNumber, req.ImageURL, req.Status, req.CreatedAt, req.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return req, nil
}

// VerifyKYC allows admin to approve/reject
func (s *KYCService) VerifyKYC(requestID, status, notes string) error {
	if status != KYCStatusVerified && status != KYCStatusRejected {
		return errors.New("invalid status")
	}

	var verifiedAt *time.Time
	if status == KYCStatusVerified {
		now := time.Now()
		verifiedAt = &now
	}

	result, err := s.DB.Exec(`
		UPDATE kyc_requests
		SET status = $1, admin_notes = $2, verified_at = $3, updated_at = $4
		WHERE id = $5
	`, status, notes, verifiedAt, time.Now(), requestID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("request not found")
	}

	return nil
}

// GetKYCStatus returns the user's current status
func (s *KYCService) GetKYCStatus(userID string) (*KYCRequest, error) {
	var req KYCRequest
	err := s.DB.QueryRow(`
		SELECT id, user_id, document_type, status, admin_notes, created_at
		FROM kyc_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&req.ID, &req.UserID, &req.DocumentType, &req.Status, &req.AdminNotes, &req.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No request found
	}
	if err != nil {
		return nil, err
	}

	return &req, nil
}
