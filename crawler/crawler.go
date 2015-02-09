package crawler

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
)

const (
	multiTagSelector = "a[href], img[src]"
	imgSelector      = "img[src]"

	srcAttr  = "src"
	hrefAttr = "href"

	crawlDepth = 1
)

// Initialize the http client with the certificates on load.
var httpClient *http.Client

func init() {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}
}

// Crawler is in charge of crawl a specific url received in a message.
// It puts new messages in the queue when it finds new urls.
// It saves images in the database.
type Crawler struct {
	db    db.Connection
	queue queue.Connection

	msg     *queue.Message
	fetcher *fetchbot.Fetcher
}

// ProcessMessage initializes a crawler to parse a specific url and crawls its html looking for images.
func ProcessMessage(q queue.Connection, d db.Connection, msg *queue.Message) {
	log.Printf("type=messageReceived msg=%v\n", msg)

	view, err := d.ViewPage(msg.JobUUID, msg.URL)
	if err != nil {
		log.Printf("type=viewPageError jobUUID=%s url=%s err=%v\n", msg.JobUUID, msg.URL, err)
		return
	}

	if !view {
		log.Printf("type=pageAlreadyViewed jobUUID=%s url=%s\n", msg.JobUUID, msg.URL)
		return
	}

	c := newCrawler(d, q, msg)
	c.fetcher = fetchbot.New(fetchbot.HandlerFunc(c.crawlResponse))
	c.fetcher.HttpClient = httpClient
	c.Crawl()
}

func newCrawler(d db.Connection, q queue.Connection, m *queue.Message) *Crawler {
	return &Crawler{
		db:    d,
		queue: q,
		msg:   m,
	}
}

// Crawl sends a GET request to the url in the message and parses the response.
// It creates new messages for new URLs in the page.
// It stores images found in the page.
func (c Crawler) Crawl() {
	c.processing()
	defer c.done()

	q := c.fetcher.Start()

	log.Printf("type=startCrawling jobUUID=%s url=%s\n", c.jobUUID(), c.msg.URL)
	q.SendStringGet(c.msg.URL)

	q.Close()
	log.Printf("type=endCrawling jobUUID=%s url=%s\n", c.jobUUID(), c.msg.URL)
}

func (c Crawler) crawlResponse(cx *fetchbot.Context, res *http.Response, err error) {
	if err != nil {
		log.Printf("type=crawlError jobUUID=%s url=%s error=%v\n", c.jobUUID(), cx.Cmd.URL(), err)
		return
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Printf("type=parseError jobUUID=%s url=%s error=%v\n", c.jobUUID(), cx.Cmd.URL(), err)
		return
	}

	c.crawlDocument(cx, doc)
}

func (c Crawler) crawlDocument(cx *fetchbot.Context, doc *goquery.Document) {
	doc.Find(multiTagSelector).Each(func(_ int, s *goquery.Selection) {
		if s.Is(imgSelector) {
			c.saveImage(cx, s)
			return
		}

		if c.continueCrawling() {
			c.enqueueURLMessage(cx, s)
		}
	})
}

func (c Crawler) saveImage(cx *fetchbot.Context, s *goquery.Selection) {
	src, ok := s.Attr(srcAttr)
	if !ok {
		log.Printf("type=unknownImageSource jobUUID=%s selector=%v\n", c.jobUUID(), s)
		return
	}

	abs, err := cx.Cmd.URL().Parse(src)
	if err != nil {
		log.Printf("type=urlParseError jobUUID=%s src=%v err=%v\n", c.jobUUID(), src, err)
		return
	}

	err = c.db.Save(c.jobUUID(), abs.String())
	if err != nil {
		log.Printf("type=saveError jobUUID=%s imageSrc=%v err=%v\n", c.jobUUID(), abs, err)
	}
}

func (c Crawler) enqueueURLMessage(cx *fetchbot.Context, s *goquery.Selection) {
	href, ok := s.Attr(hrefAttr)
	if !ok {
		log.Printf("type=unknownHref jobUUID=%s selector=%v\n", c.jobUUID(), s)
		return
	}

	abs, err := cx.Cmd.URL().Parse(href)
	if err != nil {
		log.Printf("type=urlParseError jobUUID=%s src=%v err=%v\n", c.jobUUID(), href, err)
		return
	}

	err = c.queue.Publish(c.jobUUID(), abs.String(), c.msg.Depth+1)
	if err != nil {
		log.Printf("type=publishingError jobUUID=%s msg=%v err=%v\n", c.jobUUID(), c.msg, err)
	}
}

func (c Crawler) processing() {
	err := c.db.Processing(c.jobUUID())
	if err != nil {
		log.Printf("type=incProcessingError jobUUID=%s err=%v\n", c.jobUUID(), err)
	}
}

func (c Crawler) done() {
	err := c.db.Done(c.jobUUID())
	if err != nil {
		log.Printf("type=incDoneError jobUUID=%s err=%v\n", c.jobUUID(), err)
	}
}

func (c Crawler) jobUUID() string {
	return c.msg.JobUUID
}

func (c Crawler) continueCrawling() bool {
	return c.msg.Depth < crawlDepth
}
