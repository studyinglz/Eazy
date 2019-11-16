package limiter

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestTimeLimiter(t *testing.T) {
	rate := 5
	sum := 20
	result := sum / rate
	limiter := NewTimeLimiter(0, rate)
	var wg sync.WaitGroup
	wg.Add(sum)
	before := time.Now().UnixNano()
	for i := 0; i < sum; i++ {
		go func() {
			defer wg.Done()
			limiter.Wait(false)
		}()
	}
	wg.Wait()
	after := time.Now().UnixNano()
	used := float64(after-before) / math.Pow(10, 9)
	if math.Abs(used-float64(result)) >= 1 {
		t.Error("time limit is not ok")
	}
}
