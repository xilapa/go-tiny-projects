package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	SqLite   SqLite   `yaml:"sqlite"`
	RabbitMq RabbitMq `yaml:"rabbitmq"`
}

type SqLite struct {
	ConnectionString string `env-required:"true" yaml:"connection_string" env:"SQLITE_CONNECTIONSTRING"`
}

type RabbitMq struct {
	Url string `env-required:"true" yaml:"url" env:"RABBITMQ_URL"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("../../config/config.yml", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
