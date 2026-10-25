package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"preditto/internal/response"
	"preditto/internal/service"
)

type CountryHandler struct {
	service *service.CountryService
}

func NewCountryHandler(service *service.CountryService) *CountryHandler {
	return &CountryHandler{service: service}
}

// GetAllCountries list countries.
// @Summary List countries
// @Tags Countries
// @Description No body or parameters required; filters are ignored. Ordered by name; empty results are [].
// @Produce json
// @Success 200 {object} apidocs.Success{data=apidocs.CountriesData}
// @Failure 500 {object} apidocs.Error "INTERNAL_ERROR"
// @Router /api/teams/get_all_countries [post]
func (h *CountryHandler) GetAllCountries(c *gin.Context) {
	countries, err := h.service.GetAllCountries(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, response.InternalError)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"countries": countries})
}
