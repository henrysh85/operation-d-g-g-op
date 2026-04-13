package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type PublicationsHandler struct{ Repo *repo.PublicationsRepo }

func NewPublicationsHandler(r *repo.PublicationsRepo) *PublicationsHandler {
	return &PublicationsHandler{Repo: r}
}

func (h *PublicationsHandler) List(c *gin.Context) {
	out, err := h.Repo.List(c.Request.Context(), c.Query("vertical"), qInt(c, "limit", 100))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *PublicationsHandler) Get(c *gin.Context) {
	p, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PublicationsHandler) Create(c *gin.Context) {
	var p models.Publication
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(c.Request.Context(), &p); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PublicationsHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
