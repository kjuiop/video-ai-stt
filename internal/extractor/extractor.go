package extractor

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/job"
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

func (e *Extractor) Process(ctx context.Context, videoCh <-chan *job.Job, audioCh chan<- *job.Job) error {

	wg := sync.WaitGroup{}

LOOP:
	for {
		select {
		case <-ctx.Done():
			slog.Debug("extract goroutine close")
			break LOOP

		case jobs, ok := <-videoCh:
			if !ok {
				slog.Debug("extractor videoCh closed, breaking loop")
				break LOOP
			}

			alreadyProcess := e.processed.IsProcessed(jobs.GetVideoPath(), process.EXTRACT_AUDIO_START)
			if alreadyProcess {
				continue
			}

			wg.Add(1)
			go func(jobs *job.Job) {
				defer wg.Done()

				e.processed.MarkProcessed(jobs.GetVideoPath(), process.EXTRACT_AUDIO_START)
				logger := slog.With("rid", jobs.GetRID(), "video_path", jobs.GetVideoPath())
				logger.Info("start audio extractor goroutine", "step", process.EXTRACT_AUDIO_START)

				audioPath, err := e.extractAudio(jobs)
				if err != nil {
					slog.Error("failed extract audio ffmpeg", "err", err.Error())
					return
				}

				jobs.SetAudioPath(audioPath)
				e.processed.MarkProcessed(jobs.GetVideoPath(), process.EXTRACT_AUDIO_COMPLETE)
				logger.Info("end audio extractor goroutine", "audio_path", jobs.GetAudioPath(), "step", process.EXTRACT_AUDIO_COMPLETE)
				audioCh <- jobs
			}(jobs)
		}
	}

	slog.Debug("waiting for all extract goroutines to finish")
	wg.Wait()
	slog.Debug("all extractor goroutines completed")

	return nil
}

func (e *Extractor) extractAudio(jobs *job.Job) (string, error) {

	filename := filepath.Base(jobs.GetVideoPath())
	outputPath := e.changeExtOutputPath(filepath.Join(e.cfg.OutputDir, filename))

	cmd := NewFFmpegBuilder().
		Input(jobs.GetVideoPath()).
		AudioSampleRate(e.cfg.OutputSampleRate).
		AudioChannels(1).
		MapAudio().
		UseFlacCodec().
		Output(outputPath).
		Build()

	slog.Debug("exec cmd ffmpeg", "cmd", strings.Join(cmd.Args, " "), "rid", jobs.GetRID(), "video_path", jobs.GetVideoPath(), "audio_path", jobs.GetAudioPath())

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
