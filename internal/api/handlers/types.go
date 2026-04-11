package handlers

type BasicSuccessResponse struct {
	Message string `json:"message" default:"Operation successful"`
}

type BasicErrorResponse struct {
	Error string `json:"error" default:"An error occurred"`
}
