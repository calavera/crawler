package api

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/calavera/crawler/context"
	"github.com/calavera/crawler/queue"
	"github.com/julienschmidt/httprouter"
)

const (
	jobParamName   = "jobUUID"
	defaultPort    = ":3819"
	crawlerPortKey = "CRAWLER_PORT"

	usage = `Image crawler usage:

1. Send a list of URLs to /crawl via a POST message:

$ curl -X POST -d@- http://mycrawler.com/crawl << EOF
http://www.google.com/ http://www.docker.com/
EOF

The server status is 201 after the urls are queued. The header "Location" includes the path to the status.

2. Check the status of a specific job:

$ curl -X GET http://mycrawler.com/status/aaaa-bbbb-cccc-dddd
- Processing: 2 URLs
- Done: 2 URLs
- Page views:
		- http://www.github.com -> 2 hits
		- http://www.docker.com -> 2 hits

3. Check images fetched in a specific job:

$ curl -X GET http://mycrawler.com/results/aaaa-bbbb-cccc-dddd
http://www.docker.com/static/img/bodybg.png
http://www.docker.com/static/img/logo.png
http://www.docker.com/static/img/padlock.png
`
)

// Server is the structure that controls requests to the api.
// It initializes the http handler when `StartServer` is called.
type Server struct {
	context context.Context
	router  *httprouter.Router
}

// StartServer creates a new server and initializes the http router to receive requests.
func StartServer(cx context.Context) {
	newServer(cx).start()
}

func (s *Server) start() {
	s.router.GET("/", s.index)
	s.router.POST("/crawl", s.crawl)
	s.router.GET("/status/:jobUUID", s.status)
	s.router.GET("/results/:jobUUID", s.results)

	port := serverPort()
	log.Printf("Server listening in port %s\n", port)
	log.Fatal(http.ListenAndServe(port, s.router))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, usage)
}

func (s *Server) crawl(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	urls, err := parseURLs(r)
	if err != nil || len(urls) == 0 {
		http.Error(w, "Invalid urls", http.StatusBadRequest)
		return
	}

	jobUUID, err := s.createNewJob()
	if err != nil {
		fmt.Printf("type=creatingJobError jobUUID=%s err=%v", jobUUID, err)
		http.Error(w, "Unable to create new jobs", http.StatusInternalServerError)
		return
	}

	for _, u := range urls {
		err := s.publish(jobUUID, u)
		if err != nil {
			fmt.Printf("type=publisingError jobUUID=%s url=%v err=%v", jobUUID, u, err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", fmt.Sprintf("/status/%s", jobUUID))
	w.Write([]byte(jobUUID))
}

func (s *Server) status(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jobUUID := ps.ByName(jobParamName)

	info, err := s.context.Db.Status(jobUUID)
	if err != nil {
		log.Printf("type=statusError jobUUID=%s err=%v", jobUUID, err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	b := bytes.NewBufferString(fmt.Sprintf("- Processing: %d URLs\n- Done: %d URLs\n", info.Processing, info.Done))

	pageViews := info.PageViews()
	if len(pageViews) > 0 {
		b.WriteString("- Page views:")
		for _, p := range pageViews {
			label := "hits"
			if p.Hits == 1 {
				label = "hit"
			}
			b.WriteString(fmt.Sprintf("\n\t- %s -> %d %s", p.URL, p.Hits, label))
		}
	}

	fmt.Fprintf(w, b.String())
}

func (s *Server) results(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	jobUUID := ps.ByName(jobParamName)

	images, err := s.context.Db.Results(jobUUID)
	if err != nil {
		log.Printf("type=resultsError jobUUID=%s err=%v", jobUUID, err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	b := bytes.NewBufferString("")
	for _, i := range images {
		b.Write(i)
		b.WriteString("\n")
	}

	fmt.Fprint(w, b.String())
}

func (s *Server) publish(jobUUID string, u *url.URL) error {
	return s.context.Queue.Publish(jobUUID, u.String(), 0)
}

func (s *Server) createNewJob() (string, error) {
	jobUUID := queue.UUID()
	return jobUUID, s.context.Db.CreateJob(jobUUID)
}

func newServer(cx context.Context) *Server {
	return &Server{
		context: cx,
		router:  httprouter.New(),
	}
}

func parseURLs(req *http.Request) ([]*url.URL, error) {
	var urls []*url.URL

	s := bufio.NewScanner(req.Body)
	defer req.Body.Close()
	s.Split(bufio.ScanWords)

	for s.Scan() {
		t := s.Text()

		u, err := url.Parse(t)
		if err != nil {
			log.Printf("type=parseError line=%s err=%v", t, err)
			return nil, err
		}

		urls = append(urls, u)
	}

	if err := s.Err(); err != nil {
		log.Printf("type=parseError err=%v", err)
		return nil, err
	}

	return urls, nil
}

func serverPort() string {
	if p := os.Getenv(crawlerPortKey); p != "" {
		return fmt.Sprintf(":%s", p)
	}
	return defaultPort
}
