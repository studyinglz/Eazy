package main

import (
	"eazy"
	"fmt"
	"net/http"
)

func main() {
	c := eazy.NewCollector(eazy.Async(true))
	c.OnResponseCallback(func(response *http.Response) {
		fmt.Println("I am visiting ", response.Request.URL)
	})
	c.OnHTMLCallback(".cFtMiddle a[target]", func(element *eazy.HTMLElement) {
		fmt.Println(element.Text)
		str, ok := element.DOM.Attr("href")
		if !ok {
			return
		}
		c.Search(str)
	})
	c.Search("http://www.hupu.com")
	c.Wait()
}
