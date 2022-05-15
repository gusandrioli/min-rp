package main

import (
	"fmt"
	"net/http"
	"sync"

	lb "github.com/gusandrioli/min-loadb"
)

func startServer(name string, port string) {
	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Greetings from %v\n", name)
	}))

	http.ListenAndServe(port, mux)
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(5)

	go func() {
		lb.SetReverseProxyAndServe(&lb.SetReverseProxyAndServeOpts{
			Config: &lb.Config{
				LoadBalancer: lb.LoadBalancer{Port: "8080"},
				Workers: []*lb.Worker{
					{URL: "http://localhost:8081/"},
					{URL: "http://localhost:8082/"},
					{URL: "http://localhost:8083/"},
					{URL: "http://localhost:8084/"},
					{URL: "http://localhost:8085/"},
					{URL: "http://localhost:8086/"},
					{URL: "http://localhost:8087/"},
				},
			},
		})
		wg.Done()
	}()

	go func() {
		startServer("worker1", ":8081")
		wg.Done()
	}()

	go func() {
		startServer("worker2", ":8082")
		wg.Done()
	}()

	go func() {
		startServer("worker3", ":8083")
		wg.Done()
	}()

	go func() {
		startServer("worker4", ":8084")
		wg.Done()
	}()

	go func() {
		startServer("worker5", ":8085")
		wg.Done()
	}()

	go func() {
		startServer("worker6", ":8086")
		wg.Done()
	}()

	go func() {
		startServer("worker7", ":8087")
		wg.Done()
	}()

	wg.Wait()
}
