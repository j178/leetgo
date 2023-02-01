package utils

import "time"

type RateLimiter struct {
	per  time.Duration
	last time.Time
}

func NewRateLimiter(per time.Duration) *RateLimiter {
	r := &RateLimiter{
		per:  per,
		last: time.Time{},
	}
	return r
}

func (r *RateLimiter) Take() {
	now := time.Now()
	if r.last.IsZero() {
		r.last = now
		return
	}

	sleep := r.per - now.Sub(r.last)
	if sleep > 0 {
		time.Sleep(sleep)
		r.last = now.Add(sleep)
	} else {
		r.last = now
	}
}
