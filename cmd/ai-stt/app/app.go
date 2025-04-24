package app

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/extractor"
	"video-ai-stt/internal/groq"
	"video-ai-stt/internal/job"
	"video-ai-stt/internal/process"
	"video-ai-stt/internal/watcher"
	"video-ai-stt/logger"
)

type App struct {
	cfg        *config.AISttConfig
	watcher    *watcher.Watcher
	extractor  *extractor.Extractor
	groqClient *groq.Groq
	videoCh    chan *job.Job
	audioCh    chan *job.Job
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
		cfg:        cfg,
		watcher:    watcher.NewWatcher(cfg.WatcherFiles, manager),
		extractor:  extractor.NewExtractor(cfg.Extractor, manager),
		videoCh:    make(chan *job.Job),
		audioCh:    make(chan *job.Job),
		groqClient: groq.NewGroq(cfg.Groq, manager),
	}
}

func (a *App) WatcherVideoFiles(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := a.watcher.Process(ctx, a.videoCh); err != nil {
		slog.Error("fail to watcher process", "watcher_dir", a.cfg.WatcherDir, "error", err.Error())
	}
}

func (a *App) ExtractAudio(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := a.extractor.Process(ctx, a.videoCh, a.audioCh); err != nil {
		slog.Error("fail to extractor process", "error", err.Error())
	}
}

func (a *App) GenerateSubtitle(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := a.groqClient.Process(ctx, a.audioCh); err != nil {
		slog.Error("fail to groq client process", "error", err.Error())
	}
}

func (a *App) Stop() {
}
