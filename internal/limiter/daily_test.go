package limiter

import "testing"

func TestDailyLimiterAllowAndReset(t *testing.T) {
	l := NewDailyLimiter(2)
	if !l.Allow() || !l.Allow() {
		t.Fatal("expected first two attempts allowed")
	}
	if l.Allow() {
		t.Fatal("expected third attempt denied")
	}
	l.Reset()
	if !l.Allow() {
		t.Fatal("expected allow after reset")
	}
}
