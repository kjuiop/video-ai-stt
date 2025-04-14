package app

import (
	"context"
	"log"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/logger"
)

type App struct {
	cfg *config.AISttConfig
}

func NewApplication() *App {

	cfg, err := config.LoadAISttEnvConfig()
	if err != nil {
		log.Fatalf("fail to read config err: %v", err)
	}

	if err := logger.SlogInit(cfg.Logger); err != nil {
		log.Fatalf("fail to init slog err : %v", err)
	}

	return &App{cfg: cfg}
}

func (a *App) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
}

func (a *App) Stop() {
}
