package config

import (
	env "github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
)

type Config struct {
}

func Read() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, errors.Wrap(err, "parse config from env")
	}

	return config, nil
}
