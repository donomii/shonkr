package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Shell     string
	DebugMode bool
	Slomo     bool
}

// DefaultConfig provides default configuration values
var DefaultConfig = Config{
	Shell:     "/bin/bash",
	DebugMode: false,
	Slomo:     false,
}

// Parse parses a simple configuration from bytes (basic TOML-like format)
func Parse(data []byte) (*Config, error) {
	config := DefaultConfig

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.Trim(parts[1], `"`))

		switch key {
		case "shell":
			config.Shell = value
		case "debug":
			if b, err := strconv.ParseBool(value); err == nil {
				config.DebugMode = b
			}
		case "slomo":
			if b, err := strconv.ParseBool(value); err == nil {
				config.Slomo = b
			}
		}
	}

	return &config, nil
}

// Encode encodes the configuration to bytes
func (c *Config) Encode() ([]byte, error) {
	content := fmt.Sprintf(`# Shonkr Terminal Configuration
shell = "%s"
debug = %v
slomo = %v
`, c.Shell, c.DebugMode, c.Slomo)

	return []byte(content), nil
}
