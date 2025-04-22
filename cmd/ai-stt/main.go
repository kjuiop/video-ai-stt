package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"video-ai-stt/cmd/ai-stt/app"
)

var BUILD_TIME = "no flag of BUILD_TIME"
var GIT_HASH = "no flag of GIT_HASH"
var APP_VERSION = "no flag of APP_VERSION"

func main() {

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	a := app.NewApplication()

	wg.Add(1)
	go a.WatcherVideoFiles(ctx, &wg)

	wg.Add(1)
	go a.ExtractAudio(ctx, &wg)

	wg.Add(1)
	go a.GenerateSubtitle(ctx, &wg)

	slog.Debug("ai stt app start", "git_hash", GIT_HASH, "build_time", BUILD_TIME, "app_version", APP_VERSION)

	<-exitSignal()
	a.Stop()
	cancel()
	wg.Wait()

	slog.Debug("ai stt app gracefully stopped")
}

func exitSignal() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	return sig
}
