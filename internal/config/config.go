package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	DB             `yaml:"db"`
	HTTPServer     `yaml:"server"`
	Authentication Auth `yaml:"authentication"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
}

type HTTPServer struct {
	Port string `yaml:"port"`
}

type Auth struct {
	AdminCreds Credentials `yaml:"admin-creds"`
	UserCreds  Credentials `yaml:"user-creds"`
	SecretKey  string      `yaml:"secret-key"`
}

type Credentials struct {
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, errors.New("env variable CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	yamlFileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file read error: %w", err)
	}

	var cfg Config

	err = yaml.Unmarshal(yamlFileBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config file parse error: %w", err)
	}

	return &cfg, nil
}
