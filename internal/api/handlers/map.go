package handlers

import (
	"field-scheduler-backend/internal/core/models"
	"field-scheduler-backend/internal/core/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MapHandler struct {
	parserService *services.ParserService
	blockService  services.BlockService // Future use for user-defined blocks
}

func NewMapHandler(parserService *services.ParserService, blockService services.BlockService) *MapHandler {
	return &MapHandler{parserService: parserService, blockService: blockService}
}

// GetGeneratedBlocks godoc
// @Summary List generated blocks in Lekki Phase 1
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {object} models.MapBlocksResponse
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

// GetCards godoc
// @Summary List user-defined cards (placeholder)
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {object} models.Card
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/cards [get]
func (h *MapHandler) GetCards(c *gin.Context) {
	cards, err := h.blockService.ListCards(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to fetch cards: " + err.Error()})
		return
	}
	if len(cards) == 0 {
		c.JSON(http.StatusOK, []models.Card{}) // Return empty array if no cards found
		return
	}

	c.JSON(http.StatusOK, cards)
}

// GetBlocks godoc
// @Summary List user-defined blocks
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {array} models.Block
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks [get]
func (h *MapHandler) GetBlocks(c *gin.Context) {
	blocks, err := h.blockService.ListBlocks(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to fetch blocks: " + err.Error()})
		return
	}
	if len(blocks) == 0 {
		c.JSON(http.StatusOK, []models.Block{}) // Return empty array if no blocks found
		return
	}

	c.JSON(http.StatusOK, blocks)
}

// CreateBlock godoc
// @Summary Create a new user-defined block
// @Tags map
// @Accept json
// @Produce json
// @Param block body models.CreateBlockRequest true "Block creation request"
// @Success 200 {object} models.Block
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks [post]
func (h *MapHandler) CreateBlock(c *gin.Context) {
	var req models.CreateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	block, err := h.blockService.CreateBlock(c, req.Name, req.GeometryType, req.Coordinates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to create block: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, block)
}

// UploadBlocksCSV godoc
// @Summary Upload a CSV file to sync blocks
// @Tags map
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file with block data"
// @Success 200 {object} BasicSuccessResponse
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks/upload [post]
func (h *MapHandler) UploadBlocksCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	f, _ := file.Open()
	defer f.Close()
	content, _ := io.ReadAll(f)

	err = h.parserService.SyncBlocksFromCSV(c.Request.Context(), content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "blocks synced successfully"})
}
