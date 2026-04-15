package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type RegulatoryHandler struct{ Repo *repo.RegulatoryRepo }

func NewRegulatoryHandler(r *repo.RegulatoryRepo) *RegulatoryHandler {
	return &RegulatoryHandler{Repo: r}
}

func (h *RegulatoryHandler) ListJurisdictions(c *gin.Context) {
	out, err := h.Repo.ListJurisdictions(c.Request.Context(), repo.JSFilter{
		Vertical:   c.Query("vertical"),
		CountryID:  c.Query("country_id"),
		RegionCode: c.Query("region"),
		Status:     c.Query("status"),
		Search:     c.Query("q"),
		Limit:      qInt(c, "limit", 200),
		Offset:     qInt(c, "offset", 0),
	})
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *RegulatoryHandler) GetJurisdiction(c *gin.Context) {
	j, err := h.Repo.GetJurisdiction(c.Request.Context(), c.Param("id"))
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, j)
}

func (h *RegulatoryHandler) ListCountries(c *gin.Context) {
	out, err := h.Repo.ListCountries(c.Request.Context())
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *RegulatoryHandler) ListRegions(c *gin.Context) {
	out, err := h.Repo.ListRegions(c.Request.Context())
	if err != nil {
		HandleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
