package main

import (
	"github.com/calavera/crawler/api"
	"github.com/calavera/crawler/context"
	"github.com/calavera/crawler/crawler"
)

func main() {
	c := context.NewDefaultContext()

	c.Queue.Subscribe(crawler.ProcessMessage)
	api.StartServer(c)
}
