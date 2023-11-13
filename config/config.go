package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Database `yaml:"database"`
		Scanner  `yaml:"scanner"`
	}
	Database struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		DatabaseName string `yaml:"database_name"`
		ConnPoolSize int    `yaml:"connection_pool_size"`
	}
	Scanner struct {
		InputDirectoryAbsolutePath  string `yaml:"input_directory_absolute_path"`
		OutputDirectoryAbsolutePath string `yaml:"output_directory_absolute_path"`
	}
)

func NewConfig(filepath string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
