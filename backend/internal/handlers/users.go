package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/auth"
)

type UsersHandler struct{ DB *pgxpool.Pool }

func NewUsersHandler(db *pgxpool.Pool) *UsersHandler { return &UsersHandler{DB: db} }

type userRow struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *UsersHandler) List(c *gin.Context) {
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.name, u.active, u.created_at,
		       COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		GROUP BY u.id
		ORDER BY u.email`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*userRow{}
	for rows.Next() {
		u := &userRow{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Active, &u.CreatedAt, &u.Roles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, u)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

type createUserIn struct {
	Email    string   `json:"email"    binding:"required,email"`
	Name     string   `json:"name"     binding:"required"`
	Password string   `json:"password" binding:"required,min=10"`
	Roles    []string `json:"roles"`
	Active   *bool    `json:"active"`
}

func (h *UsersHandler) Create(c *gin.Context) {
	var in createUserIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	var id string
	err = h.DB.QueryRow(c.Request.Context(),
		`INSERT INTO users (email, name, password_hash, active) VALUES ($1,$2,$3,$4) RETURNING id`,
		strings.ToLower(strings.TrimSpace(in.Email)), in.Name, string(hash), active).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "a user with that email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, role := range in.Roles {
		if _, err := h.DB.Exec(c.Request.Context(), `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, r.id FROM roles r WHERE r.name = $2
			ON CONFLICT DO NOTHING`, id, role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type patchUserIn struct {
	Name   *string   `json:"name"`
	Active *bool     `json:"active"`
	Roles  *[]string `json:"roles"`
}

func (h *UsersHandler) Patch(c *gin.Context) {
	var in patchUserIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := c.Param("id")
	// Block the only self-lockout paths: admin disabling themselves or
	// demoting themselves out of the admin role. Other admins can still do it.
	claims, _ := auth.ClaimsFrom(c)
	if claims != nil && claims.UserID == id {
		if in.Active != nil && !*in.Active {
			c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot disable your own account"})
			return
		}
		if in.Roles != nil {
			hasAdmin := false
			for _, r := range *in.Roles {
				if r == "admin" {
					hasAdmin = true
					break
				}
			}
			if !hasAdmin {
				c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot remove your own admin role"})
				return
			}
		}
	}
	if _, err := h.DB.Exec(c.Request.Context(), `
		UPDATE users SET
		  name   = COALESCE($1, name),
		  active = COALESCE($2, active),
		  updated_at = NOW()
		WHERE id = $3`, in.Name, in.Active, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if in.Roles != nil {
		if _, err := h.DB.Exec(c.Request.Context(), `DELETE FROM user_roles WHERE user_id=$1`, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, role := range *in.Roles {
			if _, err := h.DB.Exec(c.Request.Context(), `
				INSERT INTO user_roles (user_id, role_id)
				SELECT $1, r.id FROM roles r WHERE r.name = $2
				ON CONFLICT DO NOTHING`, id, role); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
	c.Status(http.StatusNoContent)
}

type changePwIn struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword"     binding:"required,min=10"`
}

// ChangePassword lets any authenticated user reset their own password by
// supplying the current one. Admins use ResetPassword (below) for others.
func (h *UsersHandler) ChangePassword(c *gin.Context) {
	var in changePwIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, ok := auth.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	var hash string
	if err := h.DB.QueryRow(c.Request.Context(),
		`SELECT password_hash FROM users WHERE id=$1`, claims.UserID).Scan(&hash); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user gone"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password incorrect"})
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.DB.Exec(c.Request.Context(),
		`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, string(newHash), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type resetPwIn struct {
	NewPassword string `json:"newPassword" binding:"required,min=10"`
}

// ResetPassword is an admin-only force-set of another user's password.
func (h *UsersHandler) ResetPassword(c *gin.Context) {
	var in resetPwIn
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ct, err := h.DB.Exec(c.Request.Context(),
		`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, string(newHash), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
