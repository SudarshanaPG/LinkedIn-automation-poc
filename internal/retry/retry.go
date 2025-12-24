package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

type Policy struct {
	Attempts  int
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Factor    float64
	Jitter    float64
	Retryable func(error) bool
	OnRetry   func(attempt int, delay time.Duration, err error)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func DefaultPolicy() Policy {
	return Policy{
		Attempts:  3,
		BaseDelay: 500 * time.Millisecond,
		MaxDelay:  6 * time.Second,
		Factor:    2,
		Jitter:    0.2,
	}
}

func Do(ctx context.Context, policy Policy, fn func(attempt int) error) error {
	if policy.Attempts <= 0 {
		policy.Attempts = 1
	}
	if policy.Factor <= 0 {
		policy.Factor = 2
	}
	if policy.BaseDelay < 0 {
		policy.BaseDelay = 0
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = policy.BaseDelay
	}
	if policy.Jitter < 0 {
		policy.Jitter = 0
	}

	for attempt := 1; attempt <= policy.Attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(attempt)
		if err == nil {
			return nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		if attempt == policy.Attempts {
			return err
		}
		if policy.Retryable != nil && !policy.Retryable(err) {
			return err
		}

		delay := backoffDelay(policy, attempt)
		if policy.OnRetry != nil {
			policy.OnRetry(attempt, delay, err)
		}
		if err := sleep(ctx, delay); err != nil {
			return err
		}
	}
	return nil
}

func backoffDelay(policy Policy, attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	delay := float64(policy.BaseDelay) * math.Pow(policy.Factor, float64(attempt-1))
	if policy.MaxDelay > 0 && delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}
	if policy.Jitter > 0 && delay > 0 {
		j := delay * policy.Jitter
		delay = delay + (rand.Float64()*2*j - j)
		if delay < 0 {
			delay = 0
		}
	}
	return time.Duration(delay)
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
