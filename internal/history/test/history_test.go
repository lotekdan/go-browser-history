package history_test

import (
	"testing"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/stretchr/testify/assert"
)

type mockBrowser struct{}

func (m *mockBrowser) GetHistory() ([]browser.Item, error) {
	return []browser.Item{{URL: "http://test.com"}}, nil
}

func TestGetAllHistory_Success(t *testing.T) {
	browsers := []browser.Browser{&mockBrowser{}}
	items, err := history.GetAllHistory(browsers)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestGetAllHistory_EmptyList(t *testing.T) {
	items, err := history.GetAllHistory([]browser.Browser{})
	assert.NoError(t, err)
	assert.Empty(t, items)
}
