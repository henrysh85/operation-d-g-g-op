package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/auth"
)

type AuthHandler struct {
	DB  *pgxpool.Pool
	Mgr *auth.Manager
}

func NewAuthHandler(db *pgxpool.Pool, mgr *auth.Manager) *AuthHandler {
	return &AuthHandler{DB: db, Mgr: mgr}
}

type loginBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var b loginBody
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var (
		id, name, hash string
		active         bool
	)
	err := h.DB.QueryRow(ctx,
		`SELECT id, name, password_hash, active FROM users WHERE LOWER(email)=LOWER($1)`, b.Email,
	).Scan(&id, &name, &hash, &active)
	if err == pgx.ErrNoRows || (err == nil && !active) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(b.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	rows, err := h.DB.Query(ctx, `
		SELECT r.name FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	roles := []string{}
	for rows.Next() {
		var rn string
		if err := rows.Scan(&rn); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		roles = append(roles, rn)
	}

	tok, err := h.Mgr.Issue(id, b.Email, roles, 12*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": tok,
		"user":  gin.H{"id": id, "email": b.Email, "name": name, "roles": roles},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := auth.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": claims.UserID, "email": claims.Email, "roles": claims.Roles, "hr_gate": claims.HRGate,
	})
}
