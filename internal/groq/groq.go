package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/job"
	"video-ai-stt/internal/process"
)

type Groq struct {
	cfg       config.Groq
	processed *process.ProcessedManager
}

func NewGroq(cfg config.Groq, processed *process.ProcessedManager) *Groq {
	return &Groq{
		cfg:       cfg,
		processed: processed,
	}
}

func (g *Groq) Process(ctx context.Context, audioCh <-chan *job.Job) error {

	wg := sync.WaitGroup{}

LOOP:
	for {
		select {
		case <-ctx.Done():
			slog.Debug("groq client goroutine close")
			break LOOP
		case jobs, ok := <-audioCh:
			if !ok {
				slog.Debug("groq client audioCh closed, breaking loop")
				break LOOP
			}
			logger := slog.With("rid", jobs.GetRID(), "video_path", jobs.GetVideoPath(), "audio_path", jobs.GetAudioPath())
			logger.Debug("groq client audioCh receive", "step", process.REQUEST_GROQ_API_START)
			alreadyProcess := g.processed.IsProcessed(jobs.GetVideoPath(), process.REQUEST_GROQ_API_START)
			if alreadyProcess {
				continue
			}

			wg.Add(1)
			go func(jobs *job.Job) {
				defer wg.Done()

				g.processed.MarkProcessed(jobs.GetVideoPath(), process.REQUEST_GROQ_API_START)
				filename, result, err := g.requestSubtitle(jobs.GetAudioPath())
				if err != nil {
					logger.Error("failed request groq api", "err", err.Error(), "step", process.REQUEST_GROQ_API_START)
					return
				}

				if err := g.generateTextFile(jobs, filename, result); err != nil {
					logger.Error("failed generate output text file", "result", result, "err", err.Error(), "step", process.REQUEST_GROQ_API_START)
					return
				}

				g.processed.MarkProcessed(jobs.GetVideoPath(), process.ALL_PROCESS_COMPLETE)
				logger.Info("end generate subtitle goroutine", "step", process.ALL_PROCESS_COMPLETE)
			}(jobs)
		}
	}

	slog.Debug("waiting for all groq goroutines to finish")
	wg.Wait()
	slog.Debug("all groq goroutines completed")

	return nil
}

func (g *Groq) requestSubtitle(audioPath string) (string, string, error) {

	// multipart/form-data 구성
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if err := writer.WriteField("model", g.cfg.STTUseModel); err != nil {
		return "", "", fmt.Errorf("failed write field model, err: %w", err)
	}

	if err := writer.WriteField("temperature", "0"); err != nil {
		return "", "", fmt.Errorf("failed write field temperature, err: %w", err)
	}

	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return "", "", fmt.Errorf("failed write field response_format, err: %w", err)
	}

	if err := writer.WriteField("timestamp_granularities[]", "word"); err != nil {
		return "", "", fmt.Errorf("failed write field timestamp_granularities, err: %w", err)
	}

	if err := writer.WriteField("language", "ko"); err != nil {
		return "", "", fmt.Errorf("failed write field language, err: %w", err)
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return "", "", fmt.Errorf("failed opening audio file: %w", err)
	}
	defer file.Close()

	filename := file.Name()

	// 파일 파트 추가
	filePart, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", "", fmt.Errorf("failed creating form file: %w", err)
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		return "", "", fmt.Errorf("failed copying audio file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed closing writer: %w", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", g.cfg.STTEndpoint, &requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed creating request: %w", err)
	}

	// 인증 및 헤더 설정
	groqAPIKey := g.cfg.APIToken
	req.Header.Set("Authorization", "Bearer "+groqAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	slog.Info("groq audio transcriptions call request", "step", process.REQUEST_GROQ_API_START)

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return filename, "", fmt.Errorf("failed sending request: %w", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed reading response: %w", err)
	}

	sttResp := STTResp{}
	if err := json.Unmarshal(body, &sttResp); err != nil {
		return "", "", fmt.Errorf("failed unmarshalling response: %w, body : %s", err, string(body))
	}

	slog.Info("groq audio transcriptions call response", "step", process.REQUEST_GROQ_API_END, "status_code", resp.StatusCode, "body", string(body), "duration", sttResp.Duration, "task", sttResp.Task, "language", sttResp.Language)
	return filename, sttResp.Text, nil
}

func (g *Groq) generateTextFile(jobs *job.Job, filename, text string) error {

	logger := slog.With("rid", jobs.GetRID(), "video_path", jobs.GetVideoPath(), "audio_path", jobs.GetAudioPath())
	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_START)

	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
	newFilename := nameWithoutExt + ".txt"
	outputPath := filepath.Join(g.cfg.OutputDir, filepath.Base(newFilename))

	if err := os.WriteFile(outputPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed creating output file: %w", err)
	}

	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_COMPLETE)

	return nil
}
