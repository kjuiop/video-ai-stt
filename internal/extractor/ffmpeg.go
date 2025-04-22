package extractor

import (
	"os/exec"
	"strconv"
)

type FFmpegBuilder struct {
	args []string
}

func NewFFmpegBuilder() *FFmpegBuilder {
	return &FFmpegBuilder{
		args: []string{"-y"},
	}
}

func (b *FFmpegBuilder) Input(inputPath string) *FFmpegBuilder {
	b.args = append(b.args, "-i", inputPath)
	return b
}

func (b *FFmpegBuilder) Output(outputPath string) *FFmpegBuilder {
	b.args = append(b.args, outputPath)
	return b
}

func (b *FFmpegBuilder) AudioSampleRate(sampleRate string) *FFmpegBuilder {
	b.args = append(b.args, "-ar", sampleRate)
	return b
}

func (b *FFmpegBuilder) AudioChannels(channels int) *FFmpegBuilder {
	b.args = append(b.args, "-ac", strconv.Itoa(channels))
	return b
}

func (b *FFmpegBuilder) MapAudio() *FFmpegBuilder {
	b.args = append(b.args, "-map", "0:a")
	return b
}

func (b *FFmpegBuilder) UseFlacCodec() *FFmpegBuilder {
	b.args = append(b.args, "-c:a", "flac")
	return b
}

func (b *FFmpegBuilder) Build() *exec.Cmd {
	return exec.Command("ffmpeg", b.args...)
}
