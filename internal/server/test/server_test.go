package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lotekdan/go-browser-history/internal/server"
	"github.com/lotekdan/go-browser-history/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockService struct{}

func (m *mockService) GetHistory() ([]service.Item, error) {
	return []service.Item{{URL: "http://test.com"}}, nil
}

func TestHistoryHandler_Success(t *testing.T) {
	srv := server.NewServer(&mockService{})
	req := httptest.NewRequest("GET", "/history", nil)
	w := httptest.NewRecorder()
	srv.HistoryHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "http://test.com")
}
