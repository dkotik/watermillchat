package watermillchat

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"time"
)

func (r *Room) cleanOut(cutoff int64, messageLimit int) {
	r.mu.Lock()
	r.messages = slices.DeleteFunc(r.messages, func(m Message) bool {
		return m.CreatedAt < cutoff
	})
	if sliceOff := len(r.messages) - messageLimit; sliceOff > 0 {
		r.messages = r.messages[sliceOff:]
	}
	r.mu.Unlock()
}

func (c *Chat) cleanup(ctx context.Context, frequency time.Duration) {
	tick := time.NewTicker(frequency)
	var t time.Time
	var roomQueue []*Room
	for {
		select {
		case <-ctx.Done():
			return
		case t = <-tick.C:
			slog.Debug("cleaning up expiring messages")
			c.mu.Lock()
			roomQueue = slices.Collect(maps.Values(c.rooms))
			c.mu.Unlock()

			for _, room := range roomQueue {
				room.cleanOut(t.Add(-c.historyRetention).Unix(), c.historyDepth)
			}
		}
	}
}
