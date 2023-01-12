package utils

import "time"

type RateLimiter struct {
	delay time.Duration
	ch    chan struct{}
}

func NewRateLimiter(delay time.Duration) *RateLimiter {
	r := &RateLimiter{
		delay: delay,
		ch:    make(chan struct{}),
	}
	go r.run()
	return r
}

func (r *RateLimiter) run() {
	for {
		if _, ok := <-r.ch; !ok {
			return
		}
		time.Sleep(r.delay)
	}
}

func (r *RateLimiter) Wait() {
	r.ch <- struct{}{}
}

func (r *RateLimiter) Stop() {
	close(r.ch)
}
