package eazy

import (
	"eazy/limiter"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"regexp"
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

type LimitRule struct {
	pattern regexp.Regexp
}
type Backend struct {
	limiter limiter.Limiter
	client  http.Client
}

func NewBackend(limiter limiter.Limiter) *Backend {
	return &Backend{
		limiter: limiter,
		client:  http.Client{},
	}
}

// It will return *response and error, response is expected to close by user
func (b *Backend) Do(r *http.Request) (*http.Response, error) {
	if err := b.limiter.Wait(false); err != nil {
		return nil, err
	}
	return b.client.Do(r)
}

// Visitor will contain logic about isVisit and set visit
type Visitor interface {
	Visited(string) bool
	SetVisit(string)
}
type VisitorInMemory struct {
	rwLock  *sync.RWMutex
	visitor map[string]bool
}

func NewVisitorMemory() *VisitorInMemory {
	return &VisitorInMemory{
		rwLock:  &sync.RWMutex{},
		visitor: make(map[string]bool),
	}
}

func (v *VisitorInMemory) Visited(str string) bool {
	v.rwLock.RLock()
	defer v.rwLock.RUnlock()
	if _, ok := v.visitor[str]; !ok {
		return false
	}
	return true
}
func (v *VisitorInMemory) SetVisit(str string) {
	v.rwLock.Lock()
	defer v.rwLock.Unlock()
	v.visitor[str] = true
}

type CollectorOption func(*Collector)

type Collector struct {
	responseCallbacks []ResponseCallback
	htmlCallbacks     []HTMLCallbackContainer
	async             bool
	wg                sync.WaitGroup
	backend           *Backend
	visitor           Visitor
}

func NewCollector(options ...CollectorOption) *Collector {
	c := Collector{
		visitor: NewVisitorMemory(),
		async:   false,
		backend:NewBackend(limiter.NewTimeLimiter(5,1000)),
	}
	for _, option := range options {
		option(&c)
	}
	return &c
}

// async will allow collector to work asynchronously
func Async(a bool) CollectorOption {
	return func(collector *Collector) {
		collector.async = a
	}
}

func AddBackend(backend *Backend) CollectorOption {
	return func(collector *Collector) {
		collector.backend = backend
	}
}

func AddVisitor(visitor *Visitor) CollectorOption {
	return func(collector *Collector) {
		collector.visitor = *visitor
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
func (c *Collector) fetch(urlS string) {
	defer c.wg.Done()
	if c.visitor.Visited(urlS) {
		return
	}
	u, err := url.Parse(urlS)
	if err != nil {
		log.Println(err)
		return
	}
	req := &http.Request{
		Method:     "GET",
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	resp, err := c.backend.Do(req)
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
	c.visitor.SetVisit(urlS)
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
