package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/service"
)

// CompanyHandler serves the company endpoints.
type CompanyHandler struct {
	companies *service.Company
}

// NewCompanyHandler wires the handler to its use cases.
func NewCompanyHandler(companies *service.Company) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

// Create handles POST /api/v1/companies.
func (h *CompanyHandler) Create(c *gin.Context) {
	var req createCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, err := h.companies.Create(c.Request.Context(), service.CreateCompanyInput{
		Name:      req.Name,
		Slug:      req.Slug,
		RepoOwner: req.RepoOwner,
		RepoName:  req.RepoName,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newCompanyResponse(company))
}

// Get handles GET /api/v1/companies/:id.
func (h *CompanyHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	company, err := h.companies.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, newCompanyResponse(company))
}

// List handles GET /api/v1/companies.
func (h *CompanyHandler) List(c *gin.Context) {
	companies, err := h.companies.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]companyResponse, 0, len(companies))
	for _, company := range companies {
		resp = append(resp, newCompanyResponse(company))
	}
	c.JSON(http.StatusOK, resp)
}
