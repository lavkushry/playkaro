package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/playkaro/backend/internal/auth"
	"github.com/playkaro/backend/internal/db"
	"github.com/playkaro/backend/internal/models"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Mobile   string `json:"mobile" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 OR username=$2)", req.Email, req.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Mobile:   req.Mobile,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert User
	err = db.DB.QueryRow(
		"INSERT INTO users (username, email, password_hash, mobile) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Username, user.Email, user.PasswordHash, user.Mobile,
	).Scan(&user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, username, email, password_hash, mobile FROM users WHERE email=$1", req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Mobile,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}
