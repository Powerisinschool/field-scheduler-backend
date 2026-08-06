package handlers

import (
	"encoding/json"
	"field-scheduler-backend/internal/core/models"
	"field-scheduler-backend/internal/core/services"
	"io"
	"mime/multipart"
	"net/http"
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
func (h *MapHandler) GetGeneratedBlocks(w http.ResponseWriter, r *http.Request) {
	// Call Python service to retrieve latest OSMnx blocks
	data, err := h.parserService.GetMapBlocks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch map data: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// GetCards godoc
// @Summary List user-defined cards (placeholder)
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {object} models.Card
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/cards [get]
func (h *MapHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.blockService.ListCards(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch cards: "+err.Error())
		return
	}
	if len(cards) == 0 {
		writeJSON(w, http.StatusOK, []models.Card{}) // Return empty array if no cards found
		return
	}

	writeJSON(w, http.StatusOK, cards)
}

// CreateCard godoc
// @Summary Create a new user-defined card
// @Tags map
// @Accept json
// @Produce json
// @Param card body CreateCardRequest true "Card creation request"
// @Success 200 {object} BasicSuccessResponse
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/cards [post]
func (h *MapHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	var req CreateCardRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid card payload: "+err.Error())
		return
	}

	// TODO: Pass the parsed card to the service layer for insertion

	writeSuccess(w, http.StatusOK, "card created successfully")
}

// DeleteCardByID godoc
// @Summary Delete a user-defined card
// @Tags map
// @Accept json
// @Produce json
// @Param id path string true "Card ID"
// @Success 200 {object} BasicSuccessResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/cards/{id} [delete]
func (h *MapHandler) DeleteCardByID(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")

	if cardID == "" {
		writeError(w, http.StatusBadRequest, "missing required card ID")
		return
	}

	// TODO: Execute service layer logic to delete the card

	writeSuccess(w, http.StatusOK, "card deleted successfully")
}

// GetBlocks godoc
// @Summary List user-defined blocks
// @Tags map
// @Accept json
// @Produce json
// @Success 200 {array} models.Block
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks [get]
func (h *MapHandler) GetBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := h.blockService.ListBlocks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch blocks: "+err.Error())
		return
	}
	if len(blocks) == 0 {
		writeJSON(w, http.StatusOK, []models.Block{}) // Return empty array if no blocks found
		return
	}

	writeJSON(w, http.StatusOK, blocks)
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
func (h *MapHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBlockRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	block, err := h.blockService.CreateBlock(r.Context(), req.Name, req.GeometryType, req.Coordinates, req.CardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create block: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, block)
}

// DeleteBlockByID godoc
// @Summary Delete a user-defined block
// @Tags map
// @Accept json
// @Produce json
// @Param id path string true "Block ID"
// @Success 200 {object} BasicSuccessResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/blocks/{id} [delete]
func (h *MapHandler) DeleteBlockByID(w http.ResponseWriter, r *http.Request) {
	blockId := r.PathValue("id")
	err := h.blockService.DeleteBlockByID(r.Context(), blockId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete block: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, "block deleted successfully")
}

// DeleteBlocksByCard godoc
// @Summary Delete all user-defined blocks associated with a card
// @Tags map
// @Accept json
// @Produce json
// @Param id path string true "Card Name"
// @Success 200 {object} BasicSuccessResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/cards/{id}/blocks [delete]
func (h *MapHandler) DeleteBlocksByCard(w http.ResponseWriter, r *http.Request) {
	cardName := r.PathValue("id")
	err := h.blockService.DeleteBlocksByCard(r.Context(), cardName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete block: "+err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, "block(s) deleted successfully")
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
func (h *MapHandler) UploadBlocksCSV(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no file uploaded")
		return
	}

	//f, _ := file.Open()
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			// do nothing
		}
	}(file)
	content, _ := io.ReadAll(file)

	err = h.parserService.SyncBlocksFromCSV(r.Context(), content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, "blocks synced successfully")
}

// ImportData godoc
// @Summary Import data into the database
// @Description Accepts a JSON payload containing arrays of cards and blocks to restore local database states.
// @Tags map
// @Accept json
// @Produce json
// @Param data body DataImportRequest true "Data import request"
// @Success 200 {object} BasicSuccessResponse
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /maps/import [post]
func (h *MapHandler) ImportData(w http.ResponseWriter, r *http.Request) {
	var req DataImportRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid import payload: "+err.Error())
		return
	}

	// TODO: Pass the parsed cards and blocks to the service layer for batch insertion

	writeSuccess(w, http.StatusOK, "data imported successfully")
}
