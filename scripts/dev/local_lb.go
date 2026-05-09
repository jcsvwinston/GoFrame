package main

import (
	"context"
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
	listenAddr := flag.String("listen", ":8090", "listener address")
	targetsCSV := flag.String("targets", "http://127.0.0.1:8091,http://127.0.0.1:8092", "comma-separated upstream URLs")
	flag.Parse()

	targets, err := parseTargets(*targetsCSV)
	if err != nil {
		log.Fatalf("parse targets: %v", err)
	}
	if len(targets) == 0 {
		log.Fatal("no upstream targets configured")
	}

	proxies := make([]*httputil.ReverseProxy, 0, len(targets))
	for _, target := range targets {
		proxies = append(proxies, buildProxy(target))
	}

	var rrCounter uint64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		index := int(atomic.AddUint64(&rrCounter, 1)-1) % len(proxies)
		proxies[index].ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:              strings.TrimSpace(*listenAddr),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("local lb listening on %s", server.Addr)
	log.Printf("upstreams: %s", strings.Join(targets, ", "))

	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("shutdown signal received: %s", sig.String())
	case err := <-errCh:
		if err != nil {
			log.Fatalf("listen error: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
		_ = server.Close()
	}
}

func parseTargets(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		target := strings.TrimSpace(part)
		if target == "" {
			continue
		}
		parsed, err := url.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("invalid target %q: %w", target, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("invalid target %q: scheme and host are required", target)
		}
		out = append(out, parsed.String())
	}
	return out, nil
}

func buildProxy(rawTarget string) *httputil.ReverseProxy {
	target, err := url.Parse(rawTarget)
	if err != nil {
		// parseTargets already validated; this should be unreachable.
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Nucleus-LB", "local-round-robin")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("upstream error target=%s method=%s path=%s err=%v", target.String(), r.Method, r.URL.Path, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	return proxy
}
