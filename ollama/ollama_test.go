package ollama_test

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/dkotik/watermillchat/ollama"
)

func newContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*5)
}

func TestBasicOllamaCall(t *testing.T) {
	bot := ollama.New("", "")
	if bot == nil {
		t.Fatal("<nil> Ollama bot")
	}

	ctx, cancel := newContext()
	defer cancel()
	answer, err := bot.SendMessage(ctx, "How are you?")
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			t.Skip("Ollama model was not found at default URL")
		}
		t.Fatal(err)
	}

	if len(answer) == 0 {
		t.Fatal("Ollama returned empty answer")
	}
	// t.Fatal(answer)
}
