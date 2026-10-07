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

func (h *CountryHandler) GetAllCountries(c *gin.Context) {
	countries, err := h.service.GetAllCountries(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		response.AbortError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, countries)
}
