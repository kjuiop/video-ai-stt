# 🎬 Video AI STT

**Video AI STT**는 업로드된 영상 파일로부터 자동으로 자막을 생성하는 서버 애플리케이션입니다.  
Groq의 Speech-to-Text(STT) API를 활용하여 빠르고 정확한 자막 생성을 지원합니다.

## 🧰 기술 스택

- **언어**: Go 1.22
- **STT 엔진**: [Groq Speech-to-Text API](https://console.groq.com/docs/speech-to-text)
- **구성 요소**:
    - `cmd/`: 애플리케이션 진입점
    - `config/`: 환경 설정 및 구성 파일
    - `internal/`: 핵심 비즈니스 로직
    - `logger/`: 로깅 유틸리티
    - `uploads/.working` : 업로드중인 영상 임시폴더
    - `uploads/` : 업로드가 완료된 영상 작업 폴더 (ai-stt 프로세스 시작)
    - `extract_audio/` : 영상에 대한 음원 추출
    - `output/` : 자막 텍스트 파일 결과물 위치
- **빌드 도구**: Makefile

## 🚀 시작하기

### 1. 저장소 클론

```bash
git clone https://github.com/kjuiop/video-ai-stt.git
cd video-ai-stt
```

### 2. 환경변수 설정

- https://groq.com/ 에서 회원가입 후 api_key_token 을 발행합니다.

```bash
export GROQ_API_KEY=your_api_key_here
```

### 3. 의존성 설치 및 빌드

```bash
go mod tidy
make build
```

### 4. 서버 실행

```bash
./video-ai-stt
```

<br />
