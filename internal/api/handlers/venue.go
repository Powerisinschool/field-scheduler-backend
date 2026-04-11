package handlers

import (
	"field-scheduler-backend/internal/core/services"

	"github.com/gin-gonic/gin"
)

type VenueHandler struct {
	service services.VenueService
}

func NewVenueHandler(service services.VenueService) *VenueHandler {
	return &VenueHandler{
		service: service,
	}
}

func (h *VenueHandler) ListVenues(c *gin.Context) {
	venues, err := h.service.ListVenues(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to retrieve venues"})
		return
	}

	c.JSON(200, venues)
}
