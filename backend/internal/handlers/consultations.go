package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type ConsultationsHandler struct{ Repo *repo.ConsultationsRepo }

func NewConsultationsHandler(r *repo.ConsultationsRepo) *ConsultationsHandler {
	return &ConsultationsHandler{Repo: r}
}

func (h *ConsultationsHandler) List(c *gin.Context) {
	var before *time.Time
	if v := c.Query("before"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			before = &t
		}
	}
	out, err := h.Repo.List(c.Request.Context(), repo.ConsultFilter{
		Vertical:   c.Query("vertical"),
		Status:     c.Query("status"),
		AssigneeID: c.Query("assignee_id"),
		Before:     before,
		Limit:      qInt(c, "limit", 200),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *ConsultationsHandler) Get(c *gin.Context) {
	cs, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cs)
}

func (h *ConsultationsHandler) Create(c *gin.Context) {
	var cs models.Consultation
	if err := c.ShouldBindJSON(&cs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(c.Request.Context(), &cs); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, cs)
}

func (h *ConsultationsHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
