package handlers

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"field-scheduler-backend/internal/core/models"
	"field-scheduler-backend/internal/core/services"
	db "field-scheduler-backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type ScheduleHandler struct {
	service          services.ScheduleService
	conductorService services.ConductorService
	parserService    *services.ParserService
	venueService     services.VenueService
}

func NewScheduleHandler(service services.ScheduleService, conductorService services.ConductorService, parserService *services.ParserService, venueService services.VenueService) *ScheduleHandler {
	return &ScheduleHandler{
		service:          service,
		conductorService: conductorService,
		parserService:    parserService,
		venueService:     venueService,
	}
}

type ListEntriesRequest struct {
	StartDate string `form:"start_date" binding:"required"` // Expecting YYYY-MM-DD
	EndDate   string `form:"end_date" binding:"required"`   // Expecting YYYY-MM-DD
}

// type ListEntriesResponse struct {
// 	Entries []models.ScheduleEntry `json:"entries"`
// }

// ListEntries godoc
// @Summary List schedule entries
// @Description Get a list of schedule entries for a given date range
// @Tags schedules
// @Accept json
// @Produce json
// @Param start_date query string true "Start date in YYYY-MM-DD format"
// @Param end_date query string true "End date in YYYY-MM-DD format"
// @Success 200 {array} models.ScheduleEntry
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /schedules [get]
func (h *ScheduleHandler) ListEntries(c *gin.Context) {
	var req ListEntriesRequest
	// Use ShouldBindQuery to parse URL parameters (e.g., ?user_id=123&schedule_date=2026-03-26)
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: err.Error()})
		return
	}

	// var userID pgtype.UUID
	// if err := userID.Scan(req.UserID); err != nil {
	// 	c.JSON(http.StatusBadRequest, "invalid uuid format")
	// 	return
	// }

	pgStartDate, err := models.ToPgDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid start date format, use YYYY-MM-DD"})
		return
	}

	pgEndDate, err := models.ToPgDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid end date format, use YYYY-MM-DD"})
		return
	}

	// Fetch from service
	entries, err := h.service.ListScheduleEntries(c.Request.Context(), nil, pgStartDate, pgEndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list schedule entries"})
		return
	}

	// Prevent returning `null` in JSON if the slice is empty
	if entries == nil {
		entries = []db.ScheduleEntry{}
	}

	conductors, err := h.conductorService.GetConductorsMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list conductors"})
		return
	}

	venues, err := h.venueService.GetVenuesMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list venues"})
		return
	}

	c.JSON(http.StatusOK, models.ToScheduleEntryModels(entries, conductors, venues))
}

// Request DTO (Data Transfer Object) for Gin JSON Validation
type CreateEntryRequest struct {
	ScheduleDate    string `json:"schedule_date" binding:"required" default:"YYYY-MM-DD"`          // Expecting YYYY-MM-DD
	StartTime       string `json:"start_time" binding:"required" default:"HH:MM:SS"`               // Expecting HH:MM:SS
	TaskDescription string `json:"task_description" binding:"required" default:"task description"` // Expecting "task description"
}

// CreateEntry godoc
// @Summary Create a schedule entry
// @Description Create a schedule entry for a specified date
// @Tags schedules
// @Accept json
// @Produce json
// @Param default body CreateEntryRequest true "Schedule entry details"
// @Success 201 {object} models.ScheduleEntry
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /schedules [post]
func (h *ScheduleHandler) CreateEntry(c *gin.Context) {
	var req CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: err.Error()})
		return
	}

	// // Parse custom types (UUIDs and Dates)
	// // userID, _ := uuid.Parse(req.UserID)
	// var userID pgtype.UUID
	// if err := userID.Scan(req.UserID); err != nil {
	// 	c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid uuid format"})
	// 	return
	// }

	date, err := models.ToPgDate(req.ScheduleDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
		return
	}

	startTime, err := models.ToPgTime(req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid time format, use HH:MM:SS"})
		return
	}

	// Pass to Service Layer
	entry, err := h.service.CreateScheduleEntry(c.Request.Context(), pgtype.UUID{}, date, startTime, req.TaskDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to create schedule entry"})
		return
	}

	conductors, err := h.conductorService.GetConductorsMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list conductors"})
		return
	}

	venues, err := h.venueService.GetVenuesMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list venues"})
		return
	}

	c.JSON(http.StatusCreated, models.ToScheduleEntryModel(entry, conductors, venues))
}

// UpdateEntry godoc
// @Summary Update a schedule entry
// @Description Update an existing schedule entry
// @Tags schedules
// @Accept json
// @Produce json
// @Param default body models.UpdateScheduleEntryParams true "Schedule entry details"
// @Success 200 {object} models.ScheduleEntry
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /schedules/:id [put]
//func (h *ScheduleHandler) UpdateEntry(c *gin.Context) {
//	var req models.UpdateScheduleEntryParams
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: err.Error()})
//		return
//	}
//
//	date, err := models.ToPgDate(req.ScheduleDate)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid date format, use YYYY-MM-DD"})
//		return
//	}
//
//	startTime, err := models.ToPgTime(req.StartTime)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, BasicErrorResponse{Error: "invalid time format, use HH:MM:SS"})
//		return
//	}
//
//	entry, err := h.service.UpdateScheduleEntry(c.Request.Context(), req.ID, date, startTime, req.TaskDescription)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to update schedule entry"})
//		return
//	}
//
//	conductors, err := h.conductorService.GetConductorsMap(c.Request.Context())
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list conductors"})
//		return
//	}
//
//	venues, err := h.venueService.GetVenuesMap(c.Request.Context())
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: "failed to list venues"})
//		return
//	}
//
//	c.JSON(http.StatusOK, models.ToScheduleEntryModel(entry, conductors, venues))
//}

// UploadPDF godoc
// @Summary Upload a schedule PDF
// @Description Upload a PDF file to sync schedule entries. Optionally specify a year to filter the entries.
// @Tags schedules
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "PDF file to upload"
// @Param year formData int false "Optional year to filter entries (e.g., 2026)"
// @Success 200 {object} BasicSuccessResponse
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /schedules/upload [post]
func (h *ScheduleHandler) UploadPDF(c *gin.Context) {
	// Get the optional year parameter
	yearStr := c.PostForm("year")
	year, _ := strconv.Atoi(yearStr) // defaults to 0 if empty or invalid

	// Extract the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close uploaded file"})
		}
	}(fileContent)

	fileBytes, err := io.ReadAll(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file contents"})
		return
	}

	// Call the service to sync the schedule
	entries, err := h.parserService.GetParsedPDF(c.Request.Context(), fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: err.Error()})
		return
	}
	err = h.service.SyncScheduleFromParsedEntries(c.Request.Context(), entries, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BasicErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BasicSuccessResponse{Message: "schedule successfully synced from PDF"})
}
