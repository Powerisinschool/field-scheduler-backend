package handlers

import (
	"field-scheduler-backend/internal/core/services"

	"github.com/gin-gonic/gin"
)

type ConductorHandler struct {
	service services.ConductorService
}

func NewConductorHandler(service services.ConductorService) *ConductorHandler {
	return &ConductorHandler{
		service: service,
	}
}

// ListConductors godoc
// @Summary List conductors
// @Description Get a list of all conductors
// @Tags conductors
// @Accept json
// @Produce json
// @Success 200 {array} models.Conductor
// @Failure 500 {object} BasicErrorResponse
// @Router /conductors [get]
func (h *ConductorHandler) ListConductors(c *gin.Context) {
	conductors, err := h.service.ListConductors(c.Request.Context())
	if err != nil {
		c.JSON(500, BasicErrorResponse{Error: "failed to retrieve conductors"})
		return
	}

	c.JSON(200, conductors)
}
