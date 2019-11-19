package main

import (
	"eazy"
	"eazy/limiter"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var counter int32 = 0
var before int32
func main() {
	backend := eazy.NewBackend(limiter.NewTimeLimiter(100, 5))
	c := eazy.NewCollector(eazy.Async(true),
		eazy.AddBackend(backend))
	c.OnResponseCallback(func(response *http.Response) {
		atomic.AddInt32(&counter, 1)
		//fmt.Println(c,":","I am visiting ", response.Request.URL)
	})
	c.OnHTMLCallback("a[target]", func(element *eazy.HTMLElement) {
		//fmt.Println(element.Text)
		str, ok := element.DOM.Attr("href")
		if !ok {
			return
		}
		c.Search(str)
	})
	c.Search("http://www.hupu.com")
	go func() {
		timer := time.Tick(time.Second)
		for {
			select {
			case <-timer:
				c:=atomic.LoadInt32(&counter)
				fmt.Println("We have send ",c-before," request last second")
				before=c
			}
		}
	}()
	c.Wait()
}
