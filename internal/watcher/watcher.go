package watcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"video-ai-stt/config"
	"video-ai-stt/internal/process"
)

type Watcher struct {
	cfg       config.WatcherFiles
	processed *process.ProcessedManager
}

func NewWatcher(cfg config.WatcherFiles, manager *process.ProcessedManager) *Watcher {
	return &Watcher{
		cfg:       cfg,
		processed: manager,
	}
}

func (w *Watcher) Process(ctx context.Context, videoCh chan<- string) error {

	slog.Debug("watcher start", "watcher_dir", w.cfg.WatcherDir)

	ticker := time.NewTicker(time.Duration(w.cfg.WatchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("close watcher file goroutine", "watcher_dir", w.cfg.WatcherDir)
			return nil
		case <-ticker.C:
			err := filepath.Walk(w.cfg.WatcherDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				filename := info.Name()
				if !w.checkVideoFile(filename) {
					return nil
				}

				alreadyProcess := w.processed.IsProcessed(path, process.WATCHER_FILE_REGISTER)
				if alreadyProcess {
					return nil
				}

				w.processed.MarkProcessed(path, process.WATCHER_FILE_REGISTER)
				videoCh <- path
				slog.Info("watcher new file", "watcher_dir", w.cfg.WatcherDir, "filename", filename, "step", process.WATCHER_FILE_REGISTER)
				return nil
			})

			if err != nil {
				slog.Error("failed watcher process file", "watcher_dir", w.cfg.WatcherDir, "error", err.Error())
			}
		}
	}
}

func (w *Watcher) checkVideoFile(filename string) bool {
	allowedExtensions := []string{".mp4", ".mkv", ".avi", ".mov"}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, extAllowed := range allowedExtensions {
		if ext == extAllowed {
			return true
		}
	}
	return false
}
