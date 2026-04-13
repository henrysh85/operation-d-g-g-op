package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type ActivitiesHandler struct{ Repo *repo.ActivitiesRepo }

func NewActivitiesHandler(r *repo.ActivitiesRepo) *ActivitiesHandler {
	return &ActivitiesHandler{Repo: r}
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
