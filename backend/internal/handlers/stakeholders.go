package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type StakeholdersHandler struct{ Repo *repo.StakeholdersRepo }

func NewStakeholdersHandler(r *repo.StakeholdersRepo) *StakeholdersHandler {
	return &StakeholdersHandler{Repo: r}
}

func (h *StakeholdersHandler) ListContacts(c *gin.Context) {
	out, err := h.Repo.ListContacts(c.Request.Context(), repo.ContactsFilter{
		Vertical: c.Query("vertical"),
		Tag:      c.Query("tag"),
		DCGGOnly: c.Query("dcgg") == "true",
		Search:   c.Query("q"),
		Limit:    qInt(c, "limit", 200),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *StakeholdersHandler) GetContact(c *gin.Context) {
	ct, err := h.Repo.GetContact(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ct)
}

func (h *StakeholdersHandler) CreateContact(c *gin.Context) {
	var ct models.Contact
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.CreateContact(c.Request.Context(), &ct); err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, ct)
}

func (h *StakeholdersHandler) DeleteContact(c *gin.Context) {
	if err := h.Repo.DeleteContact(c.Request.Context(), c.Param("id")); err != nil {
		HandleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *StakeholdersHandler) ListInstitutions(c *gin.Context) {
	out, err := h.Repo.ListInstitutions(c.Request.Context())
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

// Tree returns institutions + contacts grouped by region → country, matching the
// shape the Stakeholders view consumes.
func (h *StakeholdersHandler) Tree(c *gin.Context) {
	tree, err := h.Repo.Tree(c.Request.Context())
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tree)
}
