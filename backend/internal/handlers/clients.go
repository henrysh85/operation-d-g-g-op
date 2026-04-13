package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type ClientsHandler struct{ Repo *repo.ClientsRepo }

func NewClientsHandler(r *repo.ClientsRepo) *ClientsHandler { return &ClientsHandler{Repo: r} }

func (h *ClientsHandler) List(c *gin.Context) {
	out, err := h.Repo.List(c.Request.Context(), repo.ClientsFilter{
		Vertical: c.Query("vertical"),
		Status:   c.Query("status"),
		Limit:    qInt(c, "limit", 100),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *ClientsHandler) Get(c *gin.Context) {
	cl, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cl)
}

func (h *ClientsHandler) Create(c *gin.Context) {
	var cl models.Client
	if err := c.ShouldBindJSON(&cl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(c.Request.Context(), &cl); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, cl)
}

func (h *ClientsHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
