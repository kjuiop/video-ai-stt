package extractor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/process"
)

type Extractor struct {
	cfg       config.Extractor
	processed *process.ProcessedManager
}

func NewExtractor(cfg config.Extractor, manager *process.ProcessedManager) *Extractor {
	return &Extractor{
		cfg:       cfg,
		processed: manager,
	}
}

func (e *Extractor) Process(ctx context.Context, videoCh <-chan string, audioCh chan<- string) error {

	wg := sync.WaitGroup{}

LOOP:
	for {
		select {
		case <-ctx.Done():
			slog.Debug("extract goroutine close")
			break LOOP

		case path, ok := <-videoCh:
			if !ok {
				slog.Debug("extractor videoCh closed, breaking loop")
				break LOOP
			}

			alreadyProcess := e.processed.IsProcessed(path, process.EXTRACT_AUDIO_START)
			if alreadyProcess {
				continue
			}

			wg.Add(1)
			go func(videoPath string) {
				defer wg.Done()

				e.processed.MarkProcessed(videoPath, process.EXTRACT_AUDIO_START)
				slog.Info("start audio extractor goroutine", "path", videoPath, "step", process.EXTRACT_AUDIO_START)

				outputPath, err := e.extractAudio(videoPath)
				if err != nil {
					slog.Error("failed extract audio ffmpeg", "err", err.Error())
					return
				}

				slog.Debug("finish", "output_path", outputPath)

				audioCh <- outputPath
				e.processed.MarkProcessed(videoPath, process.EXTRACT_AUDIO_COMPLETE)
				slog.Info("end audio extractor goroutine", "path", videoPath, "step", process.EXTRACT_AUDIO_COMPLETE)
			}(path)
		}
	}

	slog.Debug("waiting for all extract goroutines to finish")
	wg.Wait()
	slog.Debug("all extractor goroutines completed")

	return nil
}

func (e *Extractor) extractAudio(path string) (string, error) {

	filename := filepath.Base(path)
	outputPath := e.changeExtOutputPath(filepath.Join(e.cfg.OutputDir, filename))

	cmd := NewFFmpegBuilder().
		Input(path).
		AudioSampleRate(e.cfg.OutputSampleRate).
		AudioChannels(1).
		MapAudio().
		UseFlacCodec().
		Output(outputPath).
		Build()

	slog.Debug("exec cmd ffmpeg", "cmd", strings.Join(cmd.Args, " "), "input_path", path, "output_path", outputPath)

	// 표준 출력 및 오류 출력 설정
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outputPath, nil
}

func (e *Extractor) changeExtOutputPath(outputPath string) string {
	ext := filepath.Ext(outputPath)
	base := strings.TrimSuffix(outputPath, ext)
	return base + e.cfg.OutputFormat
}
