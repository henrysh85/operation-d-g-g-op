package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type PeopleHandler struct{ Repo *repo.PeopleRepo }

func NewPeopleHandler(r *repo.PeopleRepo) *PeopleHandler { return &PeopleHandler{Repo: r} }

func (h *PeopleHandler) List(c *gin.Context) {
	out, err := h.Repo.List(c.Request.Context(), repo.PeopleFilter{
		Dept:   c.Query("dept"),
		Status: c.Query("status"),
		Search: c.Query("q"),
		Limit:  qInt(c, "limit", 200),
		Offset: qInt(c, "offset", 0),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *PeopleHandler) Get(c *gin.Context) {
	p, err := h.Repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PeopleHandler) Create(c *gin.Context) {
	var p models.Person
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

func (h *PeopleHandler) Patch(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	allowed := map[string]bool{
		"name": true, "email": true, "dept": true, "title": true, "reports_to": true,
		"status": true, "location": true, "country_code": true, "salary": true,
		"currency": true, "holiday_quota": true, "photo_key": true, "start_date": true,
		"metadata": true,
	}
	fields := map[string]any{}
	for k, v := range body {
		if allowed[k] {
			fields[k] = v
		}
	}
	if err := h.Repo.Patch(c.Request.Context(), c.Param("id"), fields); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PeopleHandler) Delete(c *gin.Context) {
	if err := h.Repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
