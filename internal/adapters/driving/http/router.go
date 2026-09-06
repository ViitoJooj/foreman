// Package http exposes the harness core over a Gin HTTP API under /api/v1.
// Handlers are thin: they bind the request, call a use case and map the result;
// no business logic lives here.
package http

import "github.com/gin-gonic/gin"

// Handlers bundles the HTTP handlers the router wires up.
type Handlers struct {
	Companies *CompanyHandler
	Agents    *AgentHandler
	Channels  *ChannelHandler
	Tasks     *TaskHandler
	Messages  *MessageHandler
	Commands  *CommandHandler
}

// NewRouter builds the API router. Every /api/v1 route is guarded by the
// X-Api-Key middleware.
func NewRouter(apiKey string, h Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	v1 := r.Group("/api/v1", apiKeyAuth(apiKey))

	v1.POST("/companies", h.Companies.Create)
	v1.GET("/companies", h.Companies.List)
	v1.GET("/companies/:id", h.Companies.Get)

	v1.POST("/agents", h.Agents.Create)
	v1.GET("/agents", h.Agents.List)
	v1.GET("/agents/:id", h.Agents.Get)

	v1.POST("/channels", h.Channels.Create)
	v1.GET("/channels", h.Channels.List)
	v1.GET("/channels/:id", h.Channels.Get)
	v1.POST("/channels/:id/messages", h.Messages.Create)

	v1.POST("/tasks", h.Tasks.Create)
	v1.GET("/tasks", h.Tasks.List)

	v1.POST("/commands", h.Commands.Create)
	v1.GET("/commands", h.Commands.List)
	v1.GET("/commands/:id", h.Commands.Get)

	return r
}
