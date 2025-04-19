package app

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/extractor"
	"video-ai-stt/internal/process"
	"video-ai-stt/internal/watcher"
	"video-ai-stt/logger"
)

type App struct {
	cfg       *config.AISttConfig
	watcher   *watcher.Watcher
	extractor *extractor.Extractor
	videoChan chan string
}

func NewApplication() *App {

	cfg, err := config.LoadAISttEnvConfig()
	if err != nil {
		log.Fatalf("fail to read config err: %v", err)
	}

	if err := logger.SlogInit(cfg.Logger); err != nil {
		log.Fatalf("fail to init slog err : %v", err)
	}

	manager := process.NewProcessedManager()

	return &App{
		cfg:       cfg,
		watcher:   watcher.NewWatcher(cfg.WatcherFiles, manager),
		extractor: extractor.NewExtractor(cfg.Extractor, manager),
		videoChan: make(chan string),
	}
}

func (a *App) WatcherVideoFiles(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := a.watcher.Process(ctx, a.videoChan); err != nil {
		slog.Error("fail to watcher process", "watcher_dir", a.cfg.WatcherDir, "error", err.Error())
	}
}

func (a *App) ExtractAudio(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := a.extractor.Process(ctx, a.videoChan); err != nil {
		slog.Error("fail to extractor process", "error", err.Error())
	}
}

func (a *App) Stop() {
}
