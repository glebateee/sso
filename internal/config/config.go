package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-default:"2h"`
	GRPC        gRPC          `yaml:"grpc" env-required:"true"`
}

type gRPC struct {
	Port    int           `yaml:"port" env-default:"12345"`
	Timeout time.Duration `yaml:"timeout" env-default:"2s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("path to config file is required")
	}
	if _, err := os.Stat(configPath); err != nil {
		panic("config file not found: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("error reading config file: " + err.Error())
	}
	return &cfg
}
