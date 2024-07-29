package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Storage    string `yaml:"storage"`
	HTTPServer `yaml:"http_server"`
	SecretKey  string `yaml:"secret_key"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"localhost"`
	Port    string `yaml:"port" env-default:":8080"`
}

func New() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, errors.New("env \"CONFIG_PATH\" not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file configuration is not exist: %s", err)
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, fmt.Errorf("error parse %s: %s", configPath, err)
	}

	return &config, nil
}
