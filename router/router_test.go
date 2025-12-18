package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupRouter_SwaggerEndpointExists(t *testing.T) {
	t.Parallel()

	r := SetupRouter()
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// We just assert that the route isn't 404. The swagger handler serves static UI.
	assert.NotEqual(t, http.StatusNotFound, w.Code, "swagger endpoint should be registered")
}
