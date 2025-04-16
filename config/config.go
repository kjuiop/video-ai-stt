package config

import "github.com/kelseyhightower/envconfig"

type AISttConfig struct {
	WatcherFiles
	Extractor
	Logger
}

type WatcherFiles struct {
	WatcherDir    string `envconfig:"STT_WATCHER_DIR" default:"./uploads"`
	WatcherSuffix string `envconfig:"STT_WATCHER_SUFFIX" default:"_begin"`
	WatchInterval int    `envconfig:"STT_WATCH_INTERVAL" default:"5"`
}

type Extractor struct {
	OutputDir     string `envconfig:"STT_OUTPUT_DIR" default:"./extract_audio"`
	OutputBitrate string `envconfig:"STT_OUTPUT_BITRATE" default:"96k"`
	OutputFormat  string `envconfig:"STT_OUTPUT_FORMAT" default:"mp3"`
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
