package groq

type STTResp struct {
	Task     string     `json:"task"`
	Language string     `json:"language"`
	Duration float64    `json:"duration"`
	Text     string     `json:"text"`
	Segments []Segments `json:"segments"`
	XGroq    XGroq      `json:"x_groq"`
}

type XGroq struct {
	ID string `json:"id"`
}

// Segments 자막 편집기에서 사용할
type Segments struct {
	ID         int64   `json:"id"`
	Seek       int64   `json:"seek"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	AvgLogProb float64 `json:"avg_logprob"`
}

type SegmentsSpec struct {
	// 세그먼트 고유 식별자
	ID int64 `json:"id"`
	// 오디오의 seek 위치
	Seek int64 `json:"seek"`
	// 세그먼트의 시작 위치
	Start float64 `json:"start"`
	// 세그먼트의 종료 위치
	End float64 `json:"end"`
	// 세그먼트의 텍스트
	Text string `json:"text"`
	// 샘플링 온도, (모델이 텍스트를 생성할 때의 다양성을 조절하는 매개변수
	// 높은 값일 수록 창의적이고, 다양성이며 낮은 값일 수록 예측 정확성이 높아짐
	Temperature float64 `json:"temperature"`
	// 텍스트의 평균 로그 확률
	// 낮은 로그 확률은 모델이 해당 텍스트를 예측하기 어려운 경우
	AvgLogProb float64 `json:"avg_logprob"`
	// 압축 비율, 텍스트가 원래 데이터에 비해 압축되었는지를 나타냄
	CompressionRatio float64 `json:"compression_ratio"`
	// 해당 세그먼트가 무음일 확률
	NoSpeechProb float64 `json:"no_speech_prob"`
	// 텍스트를 모델이 처리할 때 해당 텍스트를 토큰화하여 나타낸 값
	Tokens []int64 `json:"tokens"`
}

// WordSPEC 단어 단위로 잘개 쪼개짐
type WordSPEC struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}
