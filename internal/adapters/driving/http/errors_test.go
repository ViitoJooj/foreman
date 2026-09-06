package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/core/ports"
	"github.com/ViitoJooj/foreman/internal/core/service"
	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestRespondError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid task input", service.ErrInvalidTaskInput, http.StatusBadRequest},
		{"invalid message input", service.ErrInvalidMessageInput, http.StatusBadRequest},
		{"invalid company input", service.ErrInvalidCompanyInput, http.StatusBadRequest},
		{"invalid agent input", service.ErrInvalidAgentInput, http.StatusBadRequest},
		{"invalid channel input", service.ErrInvalidChannelInput, http.StatusBadRequest},
		{"filter required", service.ErrTaskFilterRequired, http.StatusBadRequest},
		{"wrapped invalid input", fmt.Errorf("ctx: %w", service.ErrInvalidTaskInput), http.StatusBadRequest},
		{"task not found", ports.ErrTaskNotFound, http.StatusNotFound},
		{"company not found", ports.ErrCompanyNotFound, http.StatusNotFound},
		{"message not found", ports.ErrMessageNotFound, http.StatusNotFound},
		{"unknown error", errors.New(testutil.RandomText()), http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			respondError(c, tc.err)

			require.Equal(t, tc.want, rec.Code)
			require.Contains(t, rec.Body.String(), `"error"`)
		})
	}
}
