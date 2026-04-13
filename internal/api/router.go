package api

import (
	"field-scheduler-backend/internal/api/handlers"
	"net/http"

	"github.com/gin-gonic/gin"

	docs "field-scheduler-backend/docs" // swagger docs

	swaggerfiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func SetupRouter(scheduleHandler *handlers.ScheduleHandler, conductorHandler *handlers.ConductorHandler, venueHandler *handlers.VenueHandler, mapHandler *handlers.MapHandler) *gin.Engine {
	router := gin.Default()

	docs.SwaggerInfo.Title = "Field Scheduler API"
	docs.SwaggerInfo.Description = "API for managing field schedules, conductors, and venues."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"

	// Simple health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI route
	docs := router.Group("/api/docs")
	{
		// This middleware catches the "/" before the wildcard can conflict with it
		docs.Use(func(c *gin.Context) {
			if c.Request.URL.Path == "/api/docs" || c.Request.URL.Path == "/api/docs/" {
				c.Redirect(http.StatusMovedPermanently, "/api/docs/index.html")
				c.Abort()
				return
			}
		})

		// Now the wildcard is the ONLY route defined in this group
		docs.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	api := router.Group("/api/v1")
	{
		schedules := api.Group("/schedules")
		{
			schedules.GET("/", scheduleHandler.ListEntries)
			schedules.POST("/", scheduleHandler.CreateEntry)

			schedules.POST("/upload", scheduleHandler.UploadPDF)
		}
		conductors := api.Group("/conductors")
		{
			conductors.GET("/", conductorHandler.ListConductors)
		}
		venues := api.Group("/venues")
		{
			venues.GET("/", venueHandler.ListVenues)
		}
		maps := api.Group("/maps")
		{
			maps.GET("/blocks", mapHandler.GetBlocks)
			maps.GET("/generated-blocks", mapHandler.GetGeneratedBlocks)
		}
	}

	return router
}
