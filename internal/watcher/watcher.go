package watcher

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"path/filepath"
	"strings"
	"video-ai-stt/config"
)

type Watcher struct {
	cfg config.WatcherFiles
}

func NewWatcher(cfg config.WatcherFiles) *Watcher {
	return &Watcher{
		cfg: cfg,
	}
}

func (w *Watcher) Process(ctx context.Context, videoCh chan<- string) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed initialize watcher, err : %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(w.cfg.WatcherDir); err != nil {
		return fmt.Errorf("failed add watcher dir : %s, err : %v", w.cfg.WatcherDir, err)
	}

	slog.Debug("watcher start", "watcher_dir", w.cfg.WatcherDir)

	for {
		select {
		case <-ctx.Done():
			slog.Debug("close watcher file goroutine", "watcher_dir", w.cfg.WatcherDir)
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				slog.Debug("close watcher event channel", "watcher_dir", w.cfg.WatcherDir)
				return nil
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				filename := filepath.Base(event.Name)
				if !w.checkVideoFile(filename) {
					continue
				}
				slog.Info("watcher new file", "watcher_dir", w.cfg.WatcherDir, "filename", filename)
				videoCh <- filepath.Join(w.cfg.WatcherDir, filename)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				slog.Debug("close watcher event error channel", "watcher_dir", w.cfg.WatcherDir)
				return nil
			}
			slog.Error("occur error watcher file", "watcher_dir", w.cfg.WatcherDir, "error", err.Error())
		}
	}
}

func (w *Watcher) checkVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov"
}
