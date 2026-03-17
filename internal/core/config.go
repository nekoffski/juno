package core

import "github.com/caarlos0/env/v11"

type Config struct {
	RestPort int `env:"JUNO_REST_PORT" envDefault:"6000"`
}

func LoadConfig() (*Config, error) {
	cfg := Config{}
	return &cfg, env.Parse(&cfg)
}
