package limiter

import (
	"errors"
	"math"
	"time"
)

var (
	NoTokenError = errors.New("no token and require async process")
)

type Limiter interface {
	Wait(bool) error
}

type TimeLimiter struct {
	time chan struct{}
}
// burst is the maximum possible token, rate is allowed added chance(which is counted by rate op/s
func NewTimeLimiter(burst int, rate int) TimeLimiter {
	l := TimeLimiter{time: make(chan struct{}, burst)}
	duration := int64(math.Pow(10, 9) / float64(rate))
	go func() {
		if rate <= 0 {
			return
		}
		t := time.Tick(time.Duration(duration))
		for {
			select {
			case <-t:
				<-l.time
			}
		}
	}()
	return l
}

func (l TimeLimiter) Wait(async bool) error {
	if async {
		select {
		case l.time <- struct{}{}:
			return nil
		default:
			return NoTokenError
		}
	} else {
		select {
		case l.time <- struct{}{}:
			return nil
		}
	}

}


type CombinedLimiter struct {
	limiters []Limiter
}
func NewCombineLimiter(l ...Limiter) CombinedLimiter {
	return CombinedLimiter{limiters: l}
}

func (c CombinedLimiter) Wait(async bool) error {
	for _, l := range c.limiters {
		if err := l.Wait(async); err != nil {
			return err
		}
	}
	return nil
}


