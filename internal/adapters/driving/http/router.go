// Package http exposes the harness core over a Gin HTTP API under /api/v1.
// Handlers are thin: they bind the request, call a use case and map the result;
// no business logic lives here.
package http

import "github.com/gin-gonic/gin"

// NewRouter builds the API router. Every /api/v1 route is guarded by the
// X-Api-Key middleware.
func NewRouter(apiKey string, tasks *TaskHandler, messages *MessageHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	v1 := r.Group("/api/v1", apiKeyAuth(apiKey))
	v1.POST("/tasks", tasks.Create)
	v1.GET("/tasks", tasks.List)
	v1.POST("/channels/:channelID/messages", messages.Create)

	return r
}
