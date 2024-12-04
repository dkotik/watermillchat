package watermillchat

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"time"
)

func (c *Chat) cleanup(ctx context.Context, frequency time.Duration) {
	tick := time.NewTicker(frequency)
	var t time.Time
	var cutoff int64
	var roomQueue []*Room
	for {
		select {
		case <-ctx.Done():
			return
		case t = <-tick.C:
			slog.Debug("cleaning up expiring messages")
			cutoff = t.Add(-c.historyRetention).Unix()
			if len(roomQueue) == 0 {
				c.mu.Lock()
				roomQueue = slices.Collect(maps.Values(c.rooms))
				c.mu.Unlock()
				time.Sleep(time.Second * 5)
			}

			for _, room := range roomQueue {
				room.mu.Lock()
				room.messages = slices.DeleteFunc(room.messages, func(m Message) bool {
					return m.CreatedAt < cutoff
				})
				if sliceOff := len(room.messages) - c.historyDepth; sliceOff > 0 {
					room.messages = room.messages[sliceOff:]
				}
				room.mu.Unlock()
			}
			roomQueue = roomQueue[:0]
		}
	}
}
