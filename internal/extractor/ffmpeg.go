package extractor

import (
	"os/exec"
	"path/filepath"
	"strings"
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

	ext := filepath.Ext(outputPath)
	base := strings.TrimSuffix(outputPath, ext)
	mp3Path := base + ".mp3"

	b.args = append(b.args, mp3Path)
	return b
}

func (b *FFmpegBuilder) AudioBitrate(bitrate string) *FFmpegBuilder {
	b.args = append(b.args, "-b:a", bitrate)
	return b
}

func (b *FFmpegBuilder) MapAudio() *FFmpegBuilder {
	b.args = append(b.args, "-map", "a")
	return b
}

func (b *FFmpegBuilder) Build() *exec.Cmd {
	return exec.Command("ffmpeg", b.args...)
}
