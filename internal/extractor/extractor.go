package extractor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"video-ai-stt/config"
)

type Extractor struct {
	cfg config.Extractor
}

func NewExtractor(cfg config.Extractor) *Extractor {
	return &Extractor{
		cfg: cfg,
	}
}

func (e *Extractor) Process(ctx context.Context, videoCh <-chan string) error {

	wg := sync.WaitGroup{}

LOOP:
	for {
		select {
		case <-ctx.Done():
			slog.Debug("context done, breaking loop")
			break LOOP

		case path, ok := <-videoCh:
			if !ok {
				slog.Debug("videoCh closed, breaking loop")
				break LOOP
			}

			wg.Add(1)
			go func(videoPath string) {
				defer wg.Done()

				slog.Info("start audio extractor goroutine", "path", videoPath)

				if err := e.extractAudio(videoPath); err != nil {
					slog.Error("failed extract audio ffmpeg", "err", err.Error())
					return
				}

				slog.Info("end audio extractor goroutine", "path", videoPath)
			}(path)
		}
	}

	slog.Debug("waiting for all extract goroutines to finish")
	wg.Wait()
	slog.Debug("all extractor goroutines completed")

	return nil
}

func (e *Extractor) extractAudio(path string) error {

	filename := filepath.Base(path)
	outputPath := filepath.Join(e.cfg.OutputDir, filename)

	cmd := NewFFmpegBuilder().
		Input(path).
		AudioBitrate(e.cfg.OutputBitrate).
		MapAudio().
		Output(outputPath).
		Build()

	slog.Debug("exec cmd ffmpeg", "cmd", strings.Join(cmd.Args, " "), "input_path", path, "output_path", outputPath)

	// 표준 출력 및 오류 출력 설정
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
