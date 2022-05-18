package lb

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

var config *Config

const (
	HTTPConnTypeTCP            string           = "tcp"
	ReverseProxyTypeRoundRobin ReverseProxyType = "RoundRobin"
	ReverseProxyTypePathPrefix ReverseProxyType = "PathPrefix"
)

type (
	Path             string // Path represents a URL path
	ReverseProxyType string
)

type Config struct {
	ReverseProxy ReverseProxy
	Workers      []*Worker
	Type         ReverseProxyType
}

type ReverseProxy struct {
	Port string
}

type Worker struct {
	URL   string
	Alive bool
	mu    sync.RWMutex
	Paths []Path
}

type SetReverseProxyAndServeOpts struct {
	*Config
}

func SetReverseProxyAndServe(opts *SetReverseProxyAndServeOpts) {
	// TODO: refactor this to receive them in form of flags
	//  e.g. go run main.go --proxy 8080 --workers 8081,8082,8083
	config = opts.Config

	go healthCheck()

	server := http.Server{
		Addr:    ":" + config.ReverseProxy.Port,
		Handler: http.HandlerFunc(ReverseProxyHandler),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}

var mu sync.Mutex
var requestCounter int = 0

func ReverseProxyHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	currentWorker := config.FindCurrentWorker(w, r)

	targetURL, err := url.Parse(currentWorker.URL)
	if err != nil {
		log.Fatal(err.Error())
	}
	requestCounter++
	mu.Unlock()

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("%v is dead.", targetURL)
		currentWorker.SetAlive(false)
		ReverseProxyHandler(w, r)
	}

	rp.ServeHTTP(w, r)
}

func (config *Config) FindCurrentWorker(w http.ResponseWriter, r *http.Request) *Worker {
	if config.Type == ReverseProxyTypePathPrefix {
		workerFoundByPath := config.FindWorkerByPath(w, r)
		if workerFoundByPath != nil && workerFoundByPath.IsAlive() {
			return workerFoundByPath
		}
	}

	// defaults to RR
	workerFoundByRR := config.FindWorkerByRoundRobin(w, r)
	if workerFoundByRR != nil && workerFoundByRR.IsAlive() {
		return workerFoundByRR
	}

	return nil
}

func (config *Config) FindWorkerByPath(w http.ResponseWriter, r *http.Request) *Worker {
	for _, worker := range config.Workers {
		if slices.Contains(worker.Paths, Path(r.URL.Path)) {
			return worker
		}
	}

	return nil
}

func (config *Config) FindWorkerByRoundRobin(w http.ResponseWriter, r *http.Request) *Worker {
	if requestCounter >= len(config.Workers) {
		requestCounter = 0
	}

	currentWorker := config.Workers[requestCounter]

	if !currentWorker.IsAlive() {
		requestCounter++
		return nil
	}

	return currentWorker
}

func (worker *Worker) SetAlive(b bool) {
	worker.mu.Lock()
	worker.Alive = b
	worker.mu.Unlock()
}

func (worker *Worker) IsAlive() bool {
	worker.mu.RLock()
	isAlive := worker.Alive
	worker.mu.RUnlock()
	return isAlive
}

func isWorkerAlive(url *url.URL) bool {
	conn, err := net.DialTimeout(HTTPConnTypeTCP, url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Cannot reach %v, error: %v", url.Host, err.Error())
		return false
	}
	defer conn.Close()

	return true
}

func healthCheck() {
	countdown := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-countdown.C:
			for _, worker := range config.Workers {
				pingURL, err := url.Parse(worker.URL)
				if err != nil {
					log.Fatal(err.Error())
				}

				a := isWorkerAlive(pingURL)

				worker.SetAlive(a)

				msg := "Alive"
				if !a {
					msg = "Dead"
				}

				log.Printf("%v is %v", worker.URL, msg)
			}
		}
	}
}
