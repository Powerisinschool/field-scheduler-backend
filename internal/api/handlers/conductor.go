package handlers

import (
	"field-scheduler-backend/internal/core/services"
	"net/http"
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
func (h *ConductorHandler) ListConductors(w http.ResponseWriter, r *http.Request) {
	conductors, err := h.service.ListConductors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve conductors")
		return
	}

	writeJSON(w, http.StatusOK, conductors)
}
