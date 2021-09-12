package config

import (
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
)

type Config struct {
	DbHost      string `env:"RSS_DB_HOST" envDefault:"localhost"`
	DbPort      int    `env:"RSS_DB_PORT" envDefault:"5444"`
	DbName      string `env:"RSS_DB_NAME" envDefault:"postgres"`
	DbUser      string `env:"RSS_DB_USER" envDefault:"postgres"`
	DbPassword  string `env:"RSS_DB_PASSWORD" envDefault:"postgres"`
	DbEnableSsl bool   `env:"RSS_DB_ENABLE_SSL" envDefault:"false"`

	ServerPort         int           `env:"RSS_SERVER_PORT" envDefault:"80"`
	ServerReadTimeout  time.Duration `env:"RSS_SERVER_READ_TIMEOUT" envDefault:"300ms"`
	ServerWriteTimeout time.Duration `env:"RSS_SERVER_WRITE_TIMEOUT" envDefault:"5000ms"`
}

func Read() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, errors.Wrap(err, "parse config from env")
	}

	return config, nil
}
