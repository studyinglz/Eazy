package eazy

import (
	"eazy/limiter"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"sync"
)

type ResponseCallback func(*http.Response)
type HTMLCallback func(*HTMLElement)
type HTMLElement struct {
	Name       string
	Text       string
	attributes []html.Attribute
	Request    *http.Request
	Response   *http.Response
	DOM        *goquery.Selection
	Index      int
}

type HTMLCallbackContainer struct {
	Selector string
	Function HTMLCallback
}
type Collector struct {
	responseCallbacks []ResponseCallback
	htmlCallbacks     []HTMLCallbackContainer
	async             bool
	wg                sync.WaitGroup
	webBackend
}
type CollectorOption func(*Collector)

type webBackend struct {
	limiter limiter.Limiter
	
}

func NewCollector(options ...CollectorOption) *Collector {
	c := Collector{}
	for _, option := range options {
		option(&c)
	}
	return &c
}
func Async(a bool) CollectorOption {
	return func(collector *Collector) {
		collector.async = a
	}
}
func NewHTMLElementFromSelectionNode(resp *http.Response, s *goquery.Selection, n *html.Node, idx int) *HTMLElement {
	return &HTMLElement{
		Name:       n.Data,
		Request:    resp.Request,
		Response:   resp,
		Text:       goquery.NewDocumentFromNode(n).Text(),
		DOM:        s,
		Index:      idx,
		attributes: n.Attr,
	}
}

func (c *Collector) handleOnHTML(resp *http.Response) error {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	for _, cc := range c.htmlCallbacks {
		i := 0
		doc.Find(cc.Selector).Each(func(_ int, s *goquery.Selection) {
			for _, n := range s.Nodes {
				e := NewHTMLElementFromSelectionNode(resp, s, n, i)
				i++
				cc.Function(e)
			}
		})
	}
	return nil
}
func (c *Collector) handleOnResponse(response *http.Response) {
	for _, fun := range c.responseCallbacks {
		fun(response)
	}
}
func (c *Collector) OnResponseCallback(callback ResponseCallback) {
	c.responseCallbacks = append(c.responseCallbacks, callback)
}
func (c *Collector) OnHTMLCallback(s string, callback HTMLCallback) {
	c.htmlCallbacks = append(c.htmlCallbacks, HTMLCallbackContainer{
		Selector: s,
		Function: callback,
	})
}
func (c *Collector) Wait() {
	c.wg.Wait()
}
func (c *Collector) Search(url string) {
	c.wg.Add(1)
	if c.async {
		go c.fetch(url)
	} else {
		c.fetch(url)
	}
}
func (c *Collector) fetch(url string) {
	//u, err := ParseFromUrl(url)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//req := &http.Request{
	//	Method:     "get",
	//	URL:        u,
	//	Proto:      "HTTP/1.1",
	//	ProtoMajor: 1,
	//	ProtoMinor: 1,
	//}
	defer c.wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	c.handleOnResponse(resp)
	err = c.handleOnHTML(resp)
	if err != nil {
		log.Println(err)
		return
	}
}
