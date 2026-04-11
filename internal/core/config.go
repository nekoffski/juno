package core

import "github.com/caarlos0/env/v11"

type DBConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Name     string `env:"POSTGRES_DB,required"`
}

type Config struct {
	RestPort         int    `env:"JUNO_REST_PORT" envDefault:"6000"`
	MetricsPort      int    `env:"JUNO_METRICS_PORT" envDefault:"6004"`
	WebPort          int    `env:"JUNO_WEB_PORT" envDefault:"6001"`
	YeelightSsdpAddr string `env:"JUNO_YEELIGHT_SSDP_ADDR" envDefault:"239.255.255.250"`
	YeelightSsdpPort int    `env:"JUNO_YEELIGHT_SSDP_PORT" envDefault:"1982"`
	LanAgentAddr     string `env:"JUNO_EDGE_ADDR" envDefault:""`
	LanAgentPort     int    `env:"JUNO_LAN_AGENT_PORT" envDefault:"7000"`
	DB               DBConfig
}

func LoadConfig() (*Config, error) {
	cfg := Config{}
	return &cfg, env.Parse(&cfg)
}
