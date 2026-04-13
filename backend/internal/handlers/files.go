package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/storage"
)

// NOTE: google/uuid is brought in transitively via pgx/v5. If the module graph
// doesn't include it directly, swap for crypto/rand-based key generation.

type FilesHandler struct{ S3 *storage.Client }

func NewFilesHandler(s *storage.Client) *FilesHandler { return &FilesHandler{S3: s} }

type presignReq struct {
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	Prefix      string `json:"prefix"`
}

func (h *FilesHandler) PresignPut(c *gin.Context) {
	var b presignReq
	_ = c.ShouldBindJSON(&b)
	key := b.Key
	if key == "" {
		prefix := b.Prefix
		if prefix == "" {
			prefix = "uploads"
		}
		key = prefix + "/" + uuid.NewString()
	}
	u, err := h.S3.PresignedPut(c.Request.Context(), key, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "url": u.String(), "expires_in": 900})
}

func (h *FilesHandler) PresignGet(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
		return
	}
	u, err := h.S3.PresignedGet(c.Request.Context(), key, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": u.String(), "expires_in": 900})
}

func (h *FilesHandler) Delete(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key required"})
		return
	}
	if err := h.S3.Remove(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
