package config

import "github.com/kelseyhightower/envconfig"

type AISttConfig struct {
	Logger
}

type Logger struct {
	Level       string `envconfig:"STT_LOG_LEVEL" default:"debug"`
	Path        string `envconfig:"STT_LOG_PATH" default:"./logs/access.log"`
	PrintStdOut bool   `envconfig:"STT_LOG_STDOUT" default:"true"`
}

func LoadAISttEnvConfig() (*AISttConfig, error) {
	var config AISttConfig
	if err := envconfig.Process("stt", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
