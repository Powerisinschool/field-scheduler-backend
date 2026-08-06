package handlers

import "field-scheduler-backend/internal/core/models"

type BasicSuccessResponse struct {
	Message string `json:"message" default:"Operation successful"`
}

type BasicErrorResponse struct {
	Error string `json:"error" default:"An error occurred"`
}

type BatchUpdateScheduleRequest struct {
	Date         string                 `json:"date"`
	Arrangements []models.ScheduleEntry `json:"arrangements"`
}

type DataImportRequest struct {
	Cards  []models.Card  `json:"cards"`
	Blocks []models.Block `json:"blocks"`
}

type CreateCardRequest struct {
	Name string `json:"name"`
}
