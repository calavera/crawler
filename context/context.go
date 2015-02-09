package context

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/apcera/nats"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
	"github.com/tpjg/goriakpbc"
)

const (
	riakAddressKey    = "CRAWLER_RIAK_URL"
	riakDockerLinkKey = "RIAK_PORT_8087_TCP"
	natsNodesKey      = "CRAWLER_GNATSD_NODES"
	natsDockerLinkKey = "GNATSD_PORT"
)

// Context holds connections to external dependencies.
// It configures the application to use specific queue and storage drivers.
type Context struct {
	Db    db.Connection    // client to talk with a database.
	Queue queue.Connection // client to talk with a queue.
}

// NewDefaultContext initializes the application context.
// It takes the Gnatsd nodes from an environment variable called CRAWLER_GNATSD_NODES, using nats://127.0.0.1:2222 by default.
// It takes Riak's address from an environment variable called CRAWLER_RIAK_URL, using 127.0.0.1:8087 by default.
func NewDefaultContext() Context {
	db := connectDb()
	qu := connectQueue(db)

	return Context{
		Db:    db,
		Queue: qu,
	}
}

// connectDb attempts to connect with a database.
// The prefered database engine is Riak, but it falls back to a in memory map if Riak is not configured.
// It exits the program if the connection fails.
func connectDb() db.Connection {
	if h, ok := ParseRiakHost(); ok {
		return ConnectRiakDb(h)
	}

	m, _ := db.NewMapConn()
	return m
}

// connectQueue attempts to connect with the cluster of Gnatsd servers.
// It exits the program if the connection fails.
// It falls back to a channel pool if the nats servers are not configured.
func connectQueue(d db.Connection) queue.Connection {
	if servers, ok := ParseNatsNodes(); ok {
		return ConnectNatsQueue(servers, d)
	}

	return queue.NewPoolConn(d)
}

// ParseRiakHost decides whether to connect the application to riak or not.
func ParseRiakHost() (string, bool) {
	if v := os.Getenv(riakAddressKey); v != "" {
		return v, true
	}
	if v := os.Getenv(riakDockerLinkKey); v != "" {
		u, err := url.Parse(v)
		if err != nil {
			log.Printf("Malformed url: %s\n", v)
			return "", false
		}
		return u.Host, true
	}
	return "", false
}

// ParseNatsNodes decies whether to connect the application to Gnatsd or not.
func ParseNatsNodes() ([]string, bool) {
	if n := os.Getenv(natsNodesKey); n != "" {
		return splitNodes(n), true
	}
	if n := os.Getenv(natsDockerLinkKey); n != "" {
		nodes := strings.Replace(n, "tcp://", "nats://", -1)
		return splitNodes(nodes), true
	}

	return nil, false
}

// ConnectRiakDb connects the application to the Riak cluster.
func ConnectRiakDb(host string) db.Connection {
	rc := riak.NewPool(host, 5)

	err := rc.Connect()
	if err != nil {
		log.Fatalf("Unable to connect to the Riak server: %v\n", err)
	}

	db, err := db.NewRiakConn(rc)
	if err != nil {
		log.Fatalf("Unable to create the storage buckets: %v\n", err)
	}

	log.Printf("Connected to Riak cluster in %s\n", host)

	return db
}

// ConnectNatsQueue connects the application to the Gnatsd cluster.
func ConnectNatsQueue(servers []string, d db.Connection) queue.Connection {
	opts := nats.DefaultOptions
	opts.Servers = servers

	nc, err := opts.Connect()
	if err != nil {
		log.Fatalf("Unable to connect to the Gnatsd servers: %v\n", err)
	}

	ec, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		log.Fatalf("Unable to encode the connection for json messages: %v\n", err)
	}

	log.Printf("Connected to Gnatsd cluster in %v\n", servers)

	return queue.NewNatsConn(d, ec)
}

func splitNodes(nodes string) []string {
	return strings.Split(strings.Replace(nodes, " ", "", -1), ",")
}
