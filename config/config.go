package config

import "github.com/kelseyhightower/envconfig"

type AISttConfig struct {
	WatcherFiles
	Extractor
	Logger
	Groq
}

type Groq struct {
	APIToken    string `envconfig:"GROQ_API_KEY" default:""`
	STTEndpoint string `envconfig:"GROQ_STT_ENDPOINT" default:"https://api.groq.com/openai/v1/audio/transcriptions"`
	STTUseModel string `envconfig:"GROQ_STT_USE_MODEL" default:"whisper-large-v3-turbo"`
	OutputDir   string `envconfig:"GROQ_OUTPUT_DIR" default:"./output"`
}

type WatcherFiles struct {
	WatcherDir    string `envconfig:"STT_WATCHER_DIR" default:"./uploads"`
	WatchInterval int    `envconfig:"STT_WATCH_INTERVAL" default:"5"`
	IgnoreDir     string `envconfig:"STT_WATCH_IGNORE_DIR" default:".working"`
}

type Extractor struct {
	OutputDir        string `envconfig:"STT_OUTPUT_DIR" default:"./extract_audio"`
	OutputSampleRate string `envconfig:"STT_OUTPUT_BITRATE" default:"16000"`
	OutputFormat     string `envconfig:"STT_OUTPUT_FORMAT" default:".flac"`
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
