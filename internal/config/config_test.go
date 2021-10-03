package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	configValues = map[string]string{
		"RSS_DB_HOST":                   "host",
		"RSS_DB_PORT":                   "5444",
		"RSS_DB_ENABLE_SSL":             "true",
		"RSS_GOOGLE_AUTH_CLIENT_ID":     "clientID",
		"RSS_GOOGLE_AUTH_CLIENT_SECRET": "secret",
	}
)

func TestRead(t *testing.T) {
	for key, val := range configValues {
		err := os.Setenv(key, val)
		assert.Nil(t, err)
	}

	cfg, err := Read()
	assert.Nil(t, err)

	assert.Equal(t, 5444, cfg.DbPort)
	assert.Equal(t, "host", cfg.DbHost)
	assert.Equal(t, "postgres", cfg.DbName)
	assert.Equal(t, 300*time.Millisecond, cfg.ServerReadTimeout)
	assert.True(t, cfg.DbEnableSsl)
}
