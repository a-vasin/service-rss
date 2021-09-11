package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	configValues = map[string]string{}
)

func TestRead(t *testing.T) {
	for key, val := range configValues {
		err := os.Setenv(key, val)
		assert.Nil(t, err)
	}

	_, err := Read()
	assert.Nil(t, err)
}
