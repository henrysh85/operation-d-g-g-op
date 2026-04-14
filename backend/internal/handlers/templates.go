package handlers

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type TemplatesHandler struct{ Repo *repo.TemplatesRepo }

func NewTemplatesHandler(r *repo.TemplatesRepo) *TemplatesHandler { return &TemplatesHandler{Repo: r} }

func (h *TemplatesHandler) List(c *gin.Context) {
	out, err := h.Repo.List(c.Request.Context(), c.Query("kind"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *TemplatesHandler) Get(c *gin.Context) {
	t, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TemplatesHandler) Create(c *gin.Context) {
	var t models.Template
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Create(c.Request.Context(), &t); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// Render previews a template against the supplied params without persisting
// anything (for the "Preview" button in the Templates UI).
func (h *TemplatesHandler) Render(c *gin.Context) {
	tpl, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	var body struct {
		Params map[string]any `json:"params"`
	}
	_ = c.ShouldBindJSON(&body)
	t, err := template.New(tpl.Slug).Parse(tpl.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template parse: " + err.Error()})
		return
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, body.Params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template exec: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rendered": buf.String()})
}

func (h *TemplatesHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
