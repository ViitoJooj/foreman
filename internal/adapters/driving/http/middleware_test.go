package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ViitoJooj/foreman/internal/testutil"
)

func TestAPIKeyAuth(t *testing.T) {
	key := testutil.RandomAPIKey()

	tests := []struct {
		name      string
		sendKey   bool
		headerVal string
		want      int
	}{
		{"valid key passes through", true, key, http.StatusOK},
		{"wrong key is rejected", true, testutil.RandomAPIKey(), http.StatusUnauthorized},
		{"missing key is rejected", false, "", http.StatusUnauthorized},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(apiKeyAuth(key))
			r.GET("/guarded", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/guarded", nil)
			if tc.sendKey {
				req.Header.Set("X-Api-Key", tc.headerVal)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.want, rec.Code)
		})
	}
}
