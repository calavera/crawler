package queue

// Message is the structure that the crawler sends and receives in the queue.
type Message struct {
	Depth   uint   // depth level where the url was found
	JobUUID string // unique identifiler for the job that trigerred this message
	URL     string // url to crawl
}

// NewMessage creates new messages to crawl an url.
func NewMessage(jobUUID, url string, d uint) *Message {
	return &Message{
		Depth:   d,
		JobUUID: jobUUID,
		URL:     url,
	}
}
