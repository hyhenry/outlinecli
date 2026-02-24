package config

import (
	"errors"
	"os"

	"github.com/henry/outlinecli/internal/credentials"
)

const defaultBaseURL = "https://app.getoutline.com/api"

type Config struct {
	APIKey  string
	BaseURL string
}

// Load resolves configuration with the following precedence:
//  1. OUTLINE_API_KEY / OUTLINE_URL environment variables
//  2. Stored credentials (~/.config/outlinecli/credentials.json)
func Load() (*Config, error) {
	cfg := &Config{}

	// 1. Environment variables (highest priority)
	cfg.APIKey = os.Getenv("OUTLINE_API_KEY")
	cfg.BaseURL = os.Getenv("OUTLINE_URL")

	// 2. Fall back to stored credentials for any missing values
	if cfg.APIKey == "" || cfg.BaseURL == "" {
		stored, err := credentials.Load()
		if err != nil {
			return nil, err
		}
		if stored != nil {
			if cfg.APIKey == "" {
				cfg.APIKey = stored.APIKey
			}
			if cfg.BaseURL == "" {
				cfg.BaseURL = stored.BaseURL
			}
		}
	}

	if cfg.APIKey == "" {
		return nil, errors.New("no API key found — run: outline auth credentials <file>")
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	return cfg, nil
}
