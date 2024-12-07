package kuura

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	LISTEN         string `env:"LISTEN" envDefault:"0.0.0.0:4000"`
	GO_ENV         string `env:"GO_ENV" envDefault:"production"`
	DATABASE_URL   string `env:"DATABASE_URL" envDefault:""`
	RUN_MIGRATIONS bool   `env:"RUN_MIGRATIONS" envDefault:"false"`
	DEBUG          bool   `env:"DEBUG" envDefault:"false"`
}

func ParseConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
