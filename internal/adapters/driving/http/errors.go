package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

// respondError maps a use-case error to an HTTP status and a JSON body. Errors
// that are not recognised are reported as a generic 500.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidTaskInput),
		errors.Is(err, service.ErrInvalidMessageInput),
		errors.Is(err, service.ErrTaskFilterRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ports.ErrTaskNotFound),
		errors.Is(err, ports.ErrCompanyNotFound),
		errors.Is(err, ports.ErrChannelNotFound),
		errors.Is(err, ports.ErrAgentNotFound),
		errors.Is(err, ports.ErrMessageNotFound),
		errors.Is(err, ports.ErrCommandNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
