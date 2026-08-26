package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

func Load[T any]() (T, error) {
	var cfg T
	if err := env.Parse(&cfg); err != nil {
		var zero T
		return zero, fmt.Errorf("parse env: %w", err)
	}
	return cfg, nil
}

func MustLoad[T any]() T {
	cfg, err := Load[T]()
	if err != nil {
		panic(err)
	}
	return cfg
}

func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d", port)
	}
	return nil
}
