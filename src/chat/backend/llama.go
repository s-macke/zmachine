package backend

import (
	"bytes"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"strings"
)

type LlamaChat struct {
	messages              []openai.ChatCompletionMessage
	totalCompletionTokens int
	totalPromptTokens     int
}

type LlamaRequest struct {
	CachePrompt bool   `json:"cache_prompt"`
	NPredict    int    `json:"n_predict"`
	Prompt      string `json:"prompt"`
}

/*
type LlamaRequest struct {
	CachePrompt      bool     `json:"cache_prompt"`
	FrequencyPenalty int      `json:"frequency_penalty"`
	Grammar          string   `json:"grammar"`
	ImageData        []any    `json:"image_data"`
	MinP             float64  `json:"min_p"`
	Mirostat         int      `json:"mirostat"`
	MirostatEta      float64  `json:"mirostat_eta"`
	MirostatTau      int      `json:"mirostat_tau"`
	NPredict         int      `json:"n_predict"`
	NProbs           int      `json:"n_probs"`
	PresencePenalty  int      `json:"presence_penalty"`
	Prompt           string   `json:"prompt"`
	RepeatLastN      int      `json:"repeat_last_n"`
	RepeatPenalty    float64  `json:"repeat_penalty"`
	SlotID           int      `json:"slot_id"`
	Stop             []string `json:"stop"`
	Stream           bool     `json:"stream"`
	Temperature      float64  `json:"temperature"`
	TfsZ             int      `json:"tfs_z"`
	TopK             int      `json:"top_k"`
	TopP             float64  `json:"top_p"`
	TypicalP         int      `json:"typical_p"`
}
*/

type LlamaResponse struct {
	Content            string `json:"content"`
	GenerationSettings struct {
		FrequencyPenalty float64       `json:"frequency_penalty"`
		Grammar          string        `json:"grammar"`
		IgnoreEos        bool          `json:"ignore_eos"`
		LogitBias        []interface{} `json:"logit_bias"`
		MinP             float64       `json:"min_p"`
		Mirostat         int           `json:"mirostat"`
		MirostatEta      float64       `json:"mirostat_eta"`
		MirostatTau      float64       `json:"mirostat_tau"`
		Model            string        `json:"model"`
		NCtx             int           `json:"n_ctx"`
		NKeep            int           `json:"n_keep"`
		NPredict         int           `json:"n_predict"`
		NProbs           int           `json:"n_probs"`
		PenalizeNl       bool          `json:"penalize_nl"`
		PresencePenalty  float64       `json:"presence_penalty"`
		RepeatLastN      int           `json:"repeat_last_n"`
		RepeatPenalty    float64       `json:"repeat_penalty"`
		Seed             int64         `json:"seed"`
		Stop             []interface{} `json:"stop"`
		Stream           bool          `json:"stream"`
		Temp             float64       `json:"temp"`
		TfsZ             float64       `json:"tfs_z"`
		TopK             int           `json:"top_k"`
		TopP             float64       `json:"top_p"`
		TypicalP         float64       `json:"typical_p"`
	} `json:"generation_settings"`
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	SlotID       int    `json:"slot_id"`
	Stop         bool   `json:"stop"`
	StoppedEos   bool   `json:"stopped_eos"`
	StoppedLimit bool   `json:"stopped_limit"`
	StoppedWord  bool   `json:"stopped_word"`
	StoppingWord string `json:"stopping_word"`
	Timings      struct {
		PredictedMs         float64 `json:"predicted_ms"`
		PredictedN          int     `json:"predicted_n"`
		PredictedPerSecond  float64 `json:"predicted_per_second"`
		PredictedPerTokenMs float64 `json:"predicted_per_token_ms"`
		PromptMs            float64 `json:"prompt_ms"`
		PromptN             int     `json:"prompt_n"`
		PromptPerSecond     float64 `json:"prompt_per_second"`
		PromptPerTokenMs    float64 `json:"prompt_per_token_ms"`
	} `json:"timings"`
	TokensCached    int  `json:"tokens_cached"`
	TokensEvaluated int  `json:"tokens_evaluated"`
	TokensPredicted int  `json:"tokens_predicted"`
	Truncated       bool `json:"truncated"`
}

func NewLlamaChat(systemMsg string) *LlamaChat {
	cs := &LlamaChat{
		totalCompletionTokens: 0,
		totalPromptTokens:     0,
	}
	cs.messages = append(cs.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemMsg,
	})
	return cs
}

func (cs *LlamaChat) PreparePrompt() string {
	var sb strings.Builder
	for _, msg := range cs.messages {
		sb.WriteString("<|im_start|>")
		sb.WriteString(msg.Role)
		sb.WriteString("\n")
		sb.WriteString(msg.Content)
		sb.WriteString("<|im_end|>\n")
	}
	sb.WriteString("<|im_start|>")
	sb.WriteString(openai.ChatMessageRoleAssistant)
	sb.WriteString("\n")
	return sb.String()
}

func (cs *LlamaChat) GetResponse(input string) (string, int, int) {
	cs.messages = append(cs.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input,
	})

	req := LlamaRequest{
		Prompt:      cs.PreparePrompt(),
		NPredict:    1024,
		CachePrompt: true,
	}

	reqAsJson, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post("http://localhost:8080/completion", "application/json", bytes.NewBuffer(reqAsJson))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(err)
	}
	defer resp.Body.Close()

	var response LlamaResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	cs.messages = append(cs.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response.Content,
	})
	return response.Content, response.TokensEvaluated, response.TokensPredicted
}
