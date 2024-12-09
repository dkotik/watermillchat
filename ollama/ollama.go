/*
Package ollama provides a generative chat bot.

Download Ollama server here: <https://ollama.com/download/mac>.
To run the model execute `ollama run ollama3.2` in terminal.
API primitives are taken from <https://dshills.medium.com/go-ollama-simple-local-ai-3a89be4bfbaf>.
*/
package ollama

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

const defaultOllamaURL = "http://localhost:11434/api/chat"

type Ollama struct {
	model      string // llama3.2
	client     http.Client
	requestURL string
	logger     *slog.Logger
}

func New(model, URL string) *Ollama {
	return &Ollama{
		model:      cmp.Or(model, "llama3.2"),
		client:     http.Client{},
		requestURL: cmp.Or(URL, defaultOllamaURL),
		logger:     slog.Default(),
	}
}

func (o *Ollama) SendMessage(ctx context.Context, m string) (string, error) {
	js, err := json.Marshal(Request{
		Model:  o.model,
		Stream: false,
		Messages: []Message{
			{
				Role:    "user",
				Content: m,
			},
			{
				Role:    "user",
				Content: "Please answer in one or two sentences.",
			},
		},
	})
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequest(http.MethodPost, o.requestURL, bytes.NewReader(js))
	if err != nil {
		return "", err
	}
	httpResp, err := o.client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	ollamaResp := Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return ollamaResp.Message.Content, err
}

func main() {
	start := time.Now()
	msg := Message{
		Role:    "user",
		Content: "Why is the sky blue?",
	}
	req := Request{
		Model:    "llama3.1",
		Stream:   false,
		Messages: []Message{msg},
	}
	resp, err := talkToOllama(defaultOllamaURL, req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(resp.Message.Content)
	fmt.Printf("Completed in %v", time.Since(start))
}

func talkToOllama(url string, ollamaReq Request) (*Response, error) {
	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	ollamaResp := Response{}
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err
}
