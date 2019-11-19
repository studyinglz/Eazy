package eazy

import (
	"eazy/limiter"
	"testing"
)

func TestAddBackend(t *testing.T) {
	backend := NewBackend(limiter.NewTimeLimiter(0, 10))
	c := NewCollector(AddBackend(backend))
	if c.backend != backend {
		t.Error("Add backend fail")
	}
}
func TestAsync(t *testing.T) {
	c := NewCollector(Async(true))
	if !c.async {
		t.Error("set async fail")
	}
}
func TestCollector_OnHTMLCallback(t *testing.T) {
	f := func(element *HTMLElement) {}
	c := NewCollector()
	c.OnHTMLCallback("a", f)
	if len(c.htmlCallbacks) == 0 {
		t.Error("add html callback  fail")

	}
}
