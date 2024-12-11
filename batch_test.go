package watermillchat_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/dkotik/watermillchat"
)

func TestBatch(t *testing.T) {
	items := make(chan int)
	go func() {
		for range 100 {
			items <- 9
			time.Sleep(time.Microsecond * time.Duration(rand.Intn(1000)))
		}
		close(items)
	}()

	out := watermillchat.Batch(items, 3, time.Millisecond)
	for batch := range out {
		for _, item := range batch {
			if item != 9 {
				t.Errorf("unexpected item: %d vs 9", item)
			}
		}
		t.Logf("end batch: %+v", batch)
	}
}
