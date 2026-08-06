package api

import (
	"field-scheduler-backend/internal/api/handlers"
	"net/http"

	docs "field-scheduler-backend/docs" // swagger docs

	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRouter(swaggerHost string, scheduleHandler *handlers.ScheduleHandler, conductorHandler *handlers.ConductorHandler, venueHandler *handlers.VenueHandler, mapHandler *handlers.MapHandler) *http.ServeMux {
	mainMux := http.NewServeMux()
	const basePath = "/api/v1"

	//router.Use(cors.Default())

	docs.SwaggerInfo.Title = "Field Scheduler API"
	docs.SwaggerInfo.Description = "API for managing field schedules, conductors, and venues."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = basePath

	mainMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status": "ok"}`))
		if err != nil {
			return
		}
	})
	mainMux.HandleFunc("GET /api/docs/", httpSwagger.Handler())

	api := func(method, pattern string, handler http.HandlerFunc) {
		fullPath := method + " " + basePath + pattern
		mainMux.HandleFunc(fullPath, handler)
	}

	api(http.MethodGet, "/schedules/", scheduleHandler.ListEntries)
	api(http.MethodPost, "/schedules/", scheduleHandler.CreateEntry)
	api(http.MethodPut, "/schedules/", scheduleHandler.BatchUpdateSchedules)
	api(http.MethodPost, "/schedules/upload/", scheduleHandler.UploadPDF)

	api(http.MethodGet, "/conductors/", conductorHandler.ListConductors)
	api(http.MethodGet, "/venues/", venueHandler.ListVenues)

	api(http.MethodGet, "/maps/cards/", mapHandler.GetCards)
	api(http.MethodPost, "/maps/cards/", mapHandler.CreateCard)
	api(http.MethodDelete, "/maps/cards/{id}/", mapHandler.DeleteCardByID)
	api(http.MethodGet, "/maps/blocks/", mapHandler.GetBlocks)
	api(http.MethodPost, "/maps/blocks/", mapHandler.CreateBlock)
	api(http.MethodDelete, "/maps/cards/{id}/blocks/", mapHandler.DeleteBlocksByCard)
	api(http.MethodDelete, "/maps/blocks/{id}/", mapHandler.DeleteBlockByID)
	api(http.MethodPost, "/maps/blocks/upload/", mapHandler.UploadBlocksCSV)
	api(http.MethodGet, "/maps/generated-blocks/", mapHandler.GetGeneratedBlocks)
	api(http.MethodPost, "/maps/import/", mapHandler.ImportData)

	return mainMux
}
