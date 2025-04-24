package process

import "sync"

const (
	WATCHER_FILE_REGISTER      = iota + 1 // 1: 작업대상파일로 등록
	EXTRACT_AUDIO_START                   // 2: 영상에서 오디오 추출 시작
	EXTRACT_AUDIO_COMPLETE                // 3: 영상에서 오디오 추출 완료
	REQUEST_GROQ_API_START                // 4: groq api request 시작
	REQUEST_GROQ_API_END                  // 5: groq api request 완료
	GENERATE_SUBTITLE_START               // 6: groq api subtitle 요청
	GENERATE_SUBTITLE_COMPLETE            // 7: groq api subtitle 완료
	ALL_PROCESS_COMPLETE                  // 8: 모든 process 완료
)

type ProcessedManager struct {
	memory *sync.Map
}

func NewProcessedManager() *ProcessedManager {
	return &ProcessedManager{memory: &sync.Map{}}
}

func (p *ProcessedManager) IsProcessed(key string, expected int) bool {
	val, ok := p.memory.Load(key)
	if !ok {
		return false
	}

	v, ok := val.(int)
	if !ok {
		return false
	}

	return v >= expected
}

func (p *ProcessedManager) MarkProcessed(key string, value int) {
	p.memory.Store(key, value)
}
