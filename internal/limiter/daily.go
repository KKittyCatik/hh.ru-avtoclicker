package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DailyLimiter struct {
	mu    sync.Mutex
	limit int
	used  int
}

func NewDailyLimiter(limit int) *DailyLimiter {
	if limit <= 0 {
		limit = 1
	}
	return &DailyLimiter{limit: limit}
}

func (d *DailyLimiter) Allow() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.used >= d.limit {
		return false
	}
	d.used++
	return true
}

func (d *DailyLimiter) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.used = 0
}

func (d *DailyLimiter) Used() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.used
}

func (d *DailyLimiter) StartAutoReset(ctx context.Context) error {
	for {
		wait := timeUntilNextMidnight(time.Now())
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return fmt.Errorf("daily limiter auto reset stopped: %w", ctx.Err())
		case <-t.C:
			d.Reset()
		}
	}
}

func timeUntilNextMidnight(now time.Time) time.Duration {
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return next.Sub(now)
}
