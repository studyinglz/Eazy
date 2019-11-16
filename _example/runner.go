package main

import (
	"eazy"
	"fmt"
	"net/http"
)

func main() {
	c := eazy.NewCollector()
	c.OnResponseCallback(func(response *http.Response) {
		fmt.Println("I am visiting ",response.Request.URL)
	})
	c.OnHTMLCallback(".cFtMiddle a[target]", func(element *eazy.HTMLElement) {
		fmt.Println(element.Text)
	})
	c.Search("http://www.hupu.com")

}
