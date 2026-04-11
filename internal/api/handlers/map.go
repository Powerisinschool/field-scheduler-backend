package handlers

import (
	"field-scheduler-backend/internal/core/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	parserService *services.ParserService
}

func NewMapHandler(parserService *services.ParserService) *MapHandler {
	return &MapHandler{parserService: parserService}
}

// GetBlocks godoc
// @Summary List blocks in Lekki Phase 1
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {object} services.MapBlocksResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks [get]
func (h *MapHandler) GetBlocks(c *gin.Context) {
	// Call Python service to retrieve latest OSMnx blocks
	data, err := h.parserService.GetMapBlocks(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to fetch map data: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
