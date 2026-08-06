package handlers

import (
	"field-scheduler-backend/internal/core/services"
	"net/http"
)

type VenueHandler struct {
	service services.VenueService
}

func NewVenueHandler(service services.VenueService) *VenueHandler {
	return &VenueHandler{
		service: service,
	}
}

// ListVenues godoc
// @Summary List venues
// @Description retrieves a list of all available venues and responds with a JSON-encoded array of venue objects.
// @Tags venue
// @Accept json
// @Produce json
// @Success 200 {array} models.Venue
// @Failure 500 {object} BasicErrorResponse
// @Router /venues [get]
func (h *VenueHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := h.service.ListVenues(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve venues")
		return
	}

	writeJSON(w, http.StatusOK, venues)
}
