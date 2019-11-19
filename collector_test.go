package eazy

import (
	"eazy/limiter"
	"net/http"
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
func TestVisitor(t *testing.T) {
	c := NewCollector()
	var counter=0
	c.OnResponseCallback(func(response *http.Response) {
		counter++
	})
	for i:=0;i<10;i++{
		c.Search("http://www.hupu.com")
	}
	if counter>1{
		t.Error("test visitor fail")
	}
}
