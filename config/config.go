package config

import (
	"errors"
	"strconv"

	"auth-login/helper"
)

type Config struct {
	AppPort   string
	JWTSecret string
}

func Load() (*Config, error) {
	settings := &Config{
		AppPort:   helper.ENV("APP_PORT"),
		JWTSecret: helper.ENV("JWT_SECRET"),
	}

	if settings.AppPort == "" {
		settings.AppPort = "8080"
	}

	port, err := strconv.Atoi(settings.AppPort)
	{
		if err != nil || port < 1 || port > 65535 {
			return nil, errors.New("APP_PORT must be between 1 and 65535")
		}
	}

	if len(settings.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	return settings, nil
}
