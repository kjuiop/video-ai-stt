package job

import "github.com/google/uuid"

type Job struct {
	rid       string
	videoPath string
	audioPath string
	filename  string
	step      int
}

func NewJob(videoPath, filename string) *Job {
	return &Job{
		rid:       uuid.NewString(),
		videoPath: videoPath,
		filename:  filename,
	}
}

func (j *Job) GetRID() string {
	return j.rid
}

func (j *Job) IsProcessed(expected int) bool {
	return j.step >= expected
}

func (j *Job) MarkProcessed(value int) {
	j.step = value
}

func (j *Job) GetVideoPath() string {
	return j.videoPath
}

func (j *Job) SetAudioPath(path string) {
	j.audioPath = path
}

func (j *Job) GetAudioPath() string {
	return j.audioPath
}
