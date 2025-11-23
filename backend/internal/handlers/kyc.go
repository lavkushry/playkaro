package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/db"
)

type UploadKYCRequest struct {
	DocumentType string `json:"document_type" binding:"required"` // PAN, AADHAAR
	DocumentURL  string `json:"document_url" binding:"required"`
}

type ApproveKYCRequest struct {
	DocumentID string `json:"document_id" binding:"required"`
	Status     string `json:"status" binding:"required"` // APPROVED, REJECTED
	Remarks    string `json:"remarks"`
}

// UploadKYCDocument allows users to submit KYC documents
func UploadKYCDocument(c *gin.Context) {
	userID := c.GetString("userID")
	var req UploadKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create KYC document record
	_, err := db.DB.Exec(
		"INSERT INTO kyc_documents (user_id, document_type, document_url, status) VALUES ($1, $2, $3, 'PENDING')",
		userID, req.DocumentType, req.DocumentURL,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document uploaded successfully. Pending review."})
}

// GetKYCStatus returns user's KYC status
func GetKYCStatus(c *gin.Context) {
	userID := c.GetString("userID")

	type KYCDoc struct {
		ID           string `json:"id"`
		DocumentType string `json:"document_type"`
		Status       string `json:"status"`
		Remarks      string `json:"remarks"`
		CreatedAt    string `json:"created_at"`
	}

	rows, err := db.DB.Query(
		"SELECT id, document_type, status, COALESCE(remarks, ''), created_at FROM kyc_documents WHERE user_id=$1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var docs []KYCDoc
	for rows.Next() {
		var doc KYCDoc
		rows.Scan(&doc.ID, &doc.DocumentType, &doc.Status, &doc.Remarks, &doc.CreatedAt)
		docs = append(docs, doc)
	}

	var kycLevel int
	db.DB.QueryRow("SELECT COALESCE(kyc_level, 0) FROM users WHERE id=$1", userID).Scan(&kycLevel)

	c.JSON(http.StatusOK, gin.H{
		"kyc_level": kycLevel,
		"documents": docs,
	})
}

// ApproveKYC allows admin to approve/reject KYC (Admin only)
func ApproveKYC(c *gin.Context) {
	adminID := c.GetString("userID")
	var req ApproveKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update document status
	var userID string
	err := db.DB.QueryRow(
		"UPDATE kyc_documents SET status=$1, reviewed_by=$2, remarks=$3 WHERE id=$4 RETURNING user_id",
		req.Status, adminID, req.Remarks, req.DocumentID,
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update KYC status"})
		return
	}

	// If approved, upgrade KYC level
	if req.Status == "APPROVED" {
		db.DB.Exec("UPDATE users SET kyc_level = kyc_level + 1 WHERE id=$1", userID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "KYC status updated"})
}
