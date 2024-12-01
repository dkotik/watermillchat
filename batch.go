package watermillchat

import "time"

// Batch periodically flushes incoming items as lists to outgoing channel.
// If list grows to limit size, it is immediately flushed.
func Batch[T any](
	in <-chan T,
	limit int,
	flush time.Duration,
) (out chan []T) {
	out = make(chan []T)

	go func() {
		tick := time.NewTicker(flush)
		batch := make([]T, 0, limit)
		for {
			select {
			case item, ok := <-in:
				if !ok {
					if len(batch) > 0 {
						batchCopy := make([]T, 0, len(batch))
						copy(batchCopy, batch)
						out <- batchCopy
					}
					close(out)
					return
				}
				batch = append(batch, item)
				if len(batch) >= limit {
					batchCopy := make([]T, 0, len(batch))
					copy(batchCopy, batch)
					out <- batchCopy
					batch = batch[:0] // truncate
				}
			case <-tick.C:
				if len(batch) > 0 {
					batchCopy := make([]T, 0, len(batch))
					copy(batchCopy, batch)
					out <- batchCopy
					batch = batch[:0] // truncate
				}
			}
		}
	}()

	return out
}
