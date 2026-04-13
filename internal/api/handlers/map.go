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

// GetGeneratedBlocks godoc
// @Summary List generated blocks in Lekki Phase 1
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {object} services.MapBlocksResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/generated-blocks [get]
func (h *MapHandler) GetGeneratedBlocks(c *gin.Context) {
	// Call Python service to retrieve latest OSMnx blocks
	data, err := h.parserService.GetMapBlocks(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to fetch map data: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *MapHandler) GetBlocks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "This endpoint will return user-defined blocks in the future."})
}
