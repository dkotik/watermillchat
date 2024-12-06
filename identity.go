package watermillchat

import "context"

type contextKeyType struct{}

var contextKey contextKeyType

type Identity struct {
	ID   string
	Name string
}

func ContextWithIdentity(parent context.Context, id Identity) context.Context {
	return context.WithValue(parent, contextKey, id)
}

func IdentityFromContext(ctx context.Context) (id Identity, ok bool) {
	id, ok = ctx.Value(contextKey).(Identity)
	return
}
