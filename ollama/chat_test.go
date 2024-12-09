package ollama_test

import "testing"

func TestConversationWithSelf(t *testing.T) {
	ctx, cancel := newContext()
	defer cancel()

	t.Log(ctx)
}
