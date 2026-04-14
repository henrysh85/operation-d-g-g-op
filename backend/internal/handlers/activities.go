package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/storage"
)

type ActivitiesHandler struct {
	Repo *repo.ActivitiesRepo
	DB   *pgxpool.Pool
	S3   *storage.Client
}

func NewActivitiesHandler(r *repo.ActivitiesRepo, db *pgxpool.Pool, s3 *storage.Client) *ActivitiesHandler {
	return &ActivitiesHandler{Repo: r, DB: db, S3: s3}
}

func (h *ActivitiesHandler) List(c *gin.Context) {
	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = &t
		}
	}
	out, err := h.Repo.List(c.Request.Context(), repo.ActivitiesFilter{
		Vertical: c.Query("vertical"),
		RegionID: c.Query("region_id"),
		ClientID: c.Query("client_id"),
		OwnerID:  c.Query("owner_id"),
		From:     from,
		To:       to,
		Search:   c.Query("q"),
		Limit:    qInt(c, "limit", 100),
		Offset:   qInt(c, "offset", 0),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *ActivitiesHandler) Get(c *gin.Context) {
	a, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ActivitiesHandler) Create(c *gin.Context) {
	var a models.Activity
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(c.Request.Context(), &a); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *ActivitiesHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadOutput accepts a multipart file, stores it in MinIO under
// outputs/<activity>/<uuid>-<filename>, and records a row in activity_outputs.
func (h *ActivitiesHandler) UploadOutput(c *gin.Context) {
	activityID := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required (multipart)"})
		return
	}
	if file.Size > 25*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "max upload is 25 MiB"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	safe := strings.ReplaceAll(file.Filename, "/", "_")
	key := fmt.Sprintf("outputs/%s/%s-%s", activityID, uuid.NewString(), safe)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := h.S3.PutObject(c.Request.Context(), key, src, file.Size, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload: " + err.Error()})
		return
	}

	var id string
	err = h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO activity_outputs (activity_id, label, minio_key, content_type, size_bytes)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		activityID, file.Filename, key, contentType, file.Size,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "register: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": id, "label": file.Filename, "minio_key": key,
		"content_type": contentType, "size_bytes": file.Size,
	})
}

// DownloadOutput streams a single activity output through the backend so the
// browser doesn't need direct access to MinIO (presigned URLs from MinIO point
// to the internal `minio:9000` endpoint and are unreachable from the public
// frontend).
func (h *ActivitiesHandler) DownloadOutput(c *gin.Context) {
	var key, label, contentType string
	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT minio_key, label, content_type
		FROM activity_outputs WHERE id = $1 AND activity_id = $2`,
		c.Param("fileId"), c.Param("id"),
	).Scan(&key, &label, &contentType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	obj, err := h.S3.GetObject(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer obj.Close()
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, label))
	c.DataFromReader(http.StatusOK, -1, contentType, obj, nil)
}

// ListOutputs returns the file metadata for an activity. The `url` field is a
// backend-proxied download URL (works from the browser regardless of MinIO
// network reachability).
func (h *ActivitiesHandler) ListOutputs(c *gin.Context) {
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id, label, minio_key, content_type, size_bytes, created_at
		FROM activity_outputs WHERE activity_id = $1 ORDER BY created_at DESC`,
		c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var (
			id, label, key, contentType string
			size                         int64
			createdAt                    time.Time
		)
		if err := rows.Scan(&id, &label, &key, &contentType, &size, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, gin.H{
			"id": id, "label": label, "minio_key": key,
			"content_type": contentType, "size_bytes": size, "created_at": createdAt,
			"url": fmt.Sprintf("/api/v1/activities/%s/outputs/%s/download", c.Param("id"), id),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *ActivitiesHandler) LinkClient(c *gin.Context) {
	var body struct {
		ClientID string `json:"client_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.LinkClient(c.Request.Context(), c.Param("id"), body.ClientID); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
