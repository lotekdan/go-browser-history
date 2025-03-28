package service_test

import (
	"testing"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockHistoryGetter struct{}

func (m *mockHistoryGetter) GetAllHistory() ([]browser.Item, error) {
	return []browser.Item{{URL: "http://test.com"}}, nil
}

func TestServiceGetHistory_Success(t *testing.T) {
	s := service.NewService(&mockHistoryGetter{})
	items, err := s.GetHistory()
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "http://test.com", items[0].URL)
}

func TestServiceGetHistory_NilGetter(t *testing.T) {
	s := service.NewService(nil)
	items, err := s.GetHistory()
	assert.Error(t, err)
	assert.Nil(t, items)
}
