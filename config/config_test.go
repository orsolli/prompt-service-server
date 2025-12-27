package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Test with environment variables set
	os.Setenv("PORT", "8080")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("CSRF_TOKEN_SECRET", "test-csrf-secret")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("CSRF_TOKEN_SECRET")
	}()

	config := LoadConfig()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, 300, config.CSRFTokenExpirySeconds)
	assert.Equal(t, "test-csrf-secret", config.CSRFTokenSecret)
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Test with no environment variables set
	config := LoadConfig()

	assert.Empty(t, config.Port)
	assert.Equal(t, 300, config.CSRFTokenExpirySeconds)
	assert.Empty(t, config.CSRFTokenSecret)
}
