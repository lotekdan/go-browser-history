package config_test

import (
	"testing"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig_DefaultPaths(t *testing.T) {
	cfg := config.NewConfig()
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.ChromePath)
	assert.NotEmpty(t, cfg.EdgePath)
	assert.NotEmpty(t, cfg.FirefoxPath)
}
