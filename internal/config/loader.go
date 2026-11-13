package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var cfg Config

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file %q: %w", path, err)
	}

	return &cfg, nil
}
