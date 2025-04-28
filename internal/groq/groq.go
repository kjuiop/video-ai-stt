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
	"strings"
	"sync"
	"video-ai-stt/config"
	"video-ai-stt/internal/job"
	"video-ai-stt/internal/process"
	"video-ai-stt/utils"
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
				filename, resp, err := g.requestSubtitle(jobs.GetAudioPath())
				if err != nil {
					logger.Error("failed request groq api", "err", err.Error(), "step", process.REQUEST_GROQ_API_START)
					return
				}

				if err := g.generateJSONFile(jobs, filename, resp); err != nil {
					logger.Error("failed generate output text file", "result", resp.Text, "err", err.Error(), "step", process.REQUEST_GROQ_API_START)
					return
				}

				if err := g.generateSRTFile(jobs, filename, resp.Segments); err != nil {
					logger.Error("failed generate output text file", "result", resp.Text, "err", err.Error(), "step", process.REQUEST_GROQ_API_START)
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

func (g *Groq) requestSubtitle(audioPath string) (string, *STTResp, error) {

	// multipart/form-data 구성
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if err := writer.WriteField("model", g.cfg.STTUseModel); err != nil {
		return "", nil, fmt.Errorf("failed write field model, err: %w", err)
	}

	if err := writer.WriteField("temperature", "0"); err != nil {
		return "", nil, fmt.Errorf("failed write field temperature, err: %w", err)
	}

	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return "", nil, fmt.Errorf("failed write field response_format, err: %w", err)
	}

	granularities := []string{"word", "segment"}
	for _, segment := range granularities {
		if err := writer.WriteField("timestamp_granularities[]", segment); err != nil {
			return "", nil, fmt.Errorf("failed write field timestamp_granularities, err: %w", err)
		}
	}

	file, err := os.Open(audioPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed opening audio file: %w", err)
	}
	defer file.Close()

	filename := file.Name()

	// 파일 파트 추가
	filePart, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", nil, fmt.Errorf("failed creating form file: %w", err)
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		return "", nil, fmt.Errorf("failed copying audio file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", nil, fmt.Errorf("failed closing writer: %w", err)
	}

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", g.cfg.STTEndpoint, &requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed creating request: %w", err)
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
		return filename, nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed curl groq api, status_code: %d, body: %s", resp.StatusCode, string(body))
	}

	sttResp := STTResp{}
	if err := json.Unmarshal(body, &sttResp); err != nil {
		return "", nil, fmt.Errorf("failed unmarshalling response: %w, body : %s", err, string(body))
	}

	slog.Info("groq audio transcriptions call response", "step", process.REQUEST_GROQ_API_END, "status_code", resp.StatusCode, "body", string(body), "duration", sttResp.Duration, "task", sttResp.Task, "language", sttResp.Language)
	return filename, &sttResp, nil
}

func (g *Groq) generateJSONFile(jobs *job.Job, filename string, resp *STTResp) error {

	outputPath := utils.GetOutputPath(g.cfg.OutputDir, filename, ".json")

	logger := slog.With("rid", jobs.GetRID(), "video_path", jobs.GetVideoPath(), "audio_path", jobs.GetAudioPath(), "json_path", outputPath, "output_path", "json")
	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_START)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed creating output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 들여쓰기를 위해 설정
	if err := encoder.Encode(resp); err != nil {
		return fmt.Errorf("failed encoding response: %w", err)
	}

	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_COMPLETE)
	return nil
}

func (g *Groq) generateSRTFile(jobs *job.Job, filename string, words []Segments) error {

	outputPath := utils.GetOutputPath(g.cfg.OutputDir, filename, ".srt")
	logger := slog.With("rid", jobs.GetRID(), "video_path", jobs.GetVideoPath(), "audio_path", jobs.GetAudioPath(), "json_path", outputPath, "output_type", "srt")
	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_START)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed creating output file: %w", err)
	}
	defer file.Close()

	srtContent := g.toSRT(words)
	_, err = file.WriteString(srtContent)
	if err != nil {
		return fmt.Errorf("failed writing to output file: %w", err)
	}

	logger.Info("generate output file", "step", process.GENERATE_SUBTITLE_COMPLETE)
	return nil
}

func srtFormatTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	milliseconds := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
}

func (g *Groq) toSRT(segments []Segments) string {
	var srt string
	for _, segment := range segments {
		start := srtFormatTime(segment.Start)
		end := srtFormatTime(segment.End)
		srt += fmt.Sprintf("%d\n%s --> %s\n%s\n\n", segment.ID+1, start, end, strings.TrimSpace(segment.Text))
	}
	return srt
}
