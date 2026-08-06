package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"field-scheduler-backend/internal/core/models"
	"field-scheduler-backend/internal/core/services"
	db "field-scheduler-backend/internal/db/sqlc"

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
	StartDate pgtype.Date `form:"start_date" binding:"required"` // Expecting YYYY-MM-DD
	EndDate   pgtype.Date `form:"end_date" binding:"required"`   // Expecting YYYY-MM-DD
}

func parseScheduleEntriesQuery(r *http.Request) (ListEntriesRequest, error) {
	query := r.URL.Query()
	var req ListEntriesRequest
	startDate := query.Get("start_date")
	endDate := query.Get("end_date")

	if startDate == "" || endDate == "" {
		return ListEntriesRequest{}, errors.New("start_date and end_date are required query parameters: ?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD")
	}

	pgStartDate, err := models.ToPgDate(startDate)
	if err != nil {
		return ListEntriesRequest{}, errors.New("invalid start date format, use YYYY-MM-DD")
	}
	req.StartDate = pgStartDate
	pgEndDate, err := models.ToPgDate(endDate)
	if err != nil {
		return ListEntriesRequest{}, errors.New("invalid end date format, use YYYY-MM-DD")
	}
	req.EndDate = pgEndDate

	return req, nil
}

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
func (h *ScheduleHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	// Use ShouldBindQuery to parse URL parameters (e.g., ?user_id=123&schedule_date=2026-03-26)
	req, err := parseScheduleEntriesQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// var userID pgtype.UUID
	// if err := userID.Scan(req.UserID); err != nil {
	// 	c.JSON(http.StatusBadRequest, "invalid uuid format")
	// 	return
	// }

	// Fetch from service
	entries, err := h.service.ListScheduleEntries(r.Context(), nil, req.StartDate, req.EndDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list schedule entries")
		return
	}

	// Prevent returning `null` in JSON if the slice is empty
	if entries == nil {
		entries = []db.ScheduleEntry{}
	}

	conductors, err := h.conductorService.GetConductorsMap(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list conductors")
		return
	}

	venues, err := h.venueService.GetVenuesMap(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list venues")
		return
	}

	writeJSON(w, http.StatusOK, models.ToScheduleEntryModels(entries, conductors, venues))
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
func (h *ScheduleHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	var req CreateEntryRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	startTime, err := models.ToPgTime(req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid time format, use HH:MM:SS")
		return
	}

	// Pass to Service Layer
	entry, err := h.service.CreateScheduleEntry(r.Context(), pgtype.UUID{}, date, startTime, req.TaskDescription)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create schedule entry")
		return
	}

	conductors, err := h.conductorService.GetConductorsMap(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list conductors")
		return
	}

	venues, err := h.venueService.GetVenuesMap(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list venues")
		return
	}

	writeJSON(w, http.StatusCreated, models.ToScheduleEntryModel(entry, conductors, venues))
}

// BatchUpdateSchedules godoc
// @Summary Batch Update a schedule entry
// @Description Update multiple schedule entries in a single request
// @Tags schedules
// @Accept json
// @Produce json
// @Param default body BatchUpdateScheduleRequest true "Schedule entry details"
// @Success 200 {object} BasicSuccessResponse
// @Failure 400 {object} BasicErrorResponse
// @Failure 500 {object} BasicErrorResponse
// @Router /schedules [put]
func (h *ScheduleHandler) BatchUpdateSchedules(w http.ResponseWriter, r *http.Request) {
	var req BatchUpdateScheduleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
		return
	}

	// TODO: Pass the parsed entries to the service layer for batch update
	//err := h.service.BatchUpdate(r.Context(), req.Date, req.Arrangements)
	//if err != nil {
	//	writeError(w, http.StatusInternalServerError, "failed to update schedule entries")
	//	return
	//}

	w.WriteHeader(http.StatusOK)
}

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
func (h *ScheduleHandler) UploadPDF(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	// Get the optional year parameter
	yearStr := r.FormValue("year")
	year, _ := strconv.Atoi(yearStr) // defaults to 0 if empty or invalid

	// Extract the file from the request
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no file uploaded")
		return
	}

	// Open the file
	//fileContent, err := file.Open()
	//if err != nil {
	//	writeError(w, http.StatusInternalServerError, "failed to open uploaded file")
	//	return
	//}
	//defer func(fileContent multipart.File) {
	//	err := fileContent.Close()
	//	if err != nil {
	//		writeError(w, http.StatusInternalServerError, "failed to close uploaded file")
	//	}
	//}(fileContent)

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			// do nothing
		}
	}(file)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file contents")
		return
	}

	// Call the service to sync the schedule
	entries, err := h.parserService.GetParsedPDF(r.Context(), fileBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = h.service.SyncScheduleFromParsedEntries(r.Context(), entries, year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, BasicSuccessResponse{Message: "schedule successfully synced from PDF"})
}
