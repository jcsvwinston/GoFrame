// cmd/balancer/main.go
// Minimal round-robin reverse proxy for local multi-node testing.
// Usage: go run cmd/balancer/main.go [upstream1] [upstream2] ...
// Default: balances between http://localhost:8091 and http://localhost:8092 on :8090
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", ":8090", "Listen address for the load balancer")
	flag.Parse()

	upstreamList := flag.Args()
	if len(upstreamList) == 0 {
		upstreamList = []string{
			"http://localhost:8091",
			"http://localhost:8092",
		}
	}

	proxies := make([]*httputil.ReverseProxy, 0, len(upstreamList))
	for _, raw := range upstreamList {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil {
			log.Fatalf("Invalid upstream %q: %v", raw, err)
		}
		p := httputil.NewSingleHostReverseProxy(u)
		p.ErrorLog = log.New(os.Stderr, fmt.Sprintf("[proxy→%s] ", u.Host), 0)
		proxies = append(proxies, p)
		log.Printf("  upstream: %s", u)
	}

	total := uint64(len(proxies))
	var counter atomic.Uint64

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var idx uint64

		// Sticky sessions: prefer the node recorded in the gf-node cookie
		if c, err := r.Cookie("gf-node"); err == nil {
			for i, raw := range upstreamList {
				if raw == c.Value {
					idx = uint64(i)
					proxies[idx].ServeHTTP(w, r)
					return
				}
			}
		}

		// New visitor: assign next node in round-robin and set sticky cookie
		idx = counter.Add(1) % total
		http.SetCookie(w, &http.Cookie{
			Name:     "gf-node",
			Value:    upstreamList[idx],
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		proxies[idx].ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("GoFrame LB (sticky) on http://localhost%s → %d nodes", *addr, total)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down balancer...")
}
