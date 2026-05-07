// cmd/ministore/main.go
// Embedded Redis-compatible server for local multi-node cluster testing.
// Provides a shared in-memory store for: sessions, admin cluster pubsub, and job queues.
// Starts a miniredis server on :6379 (or $MINISTORE_PORT) and blocks until SIGINT/SIGTERM.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alicebob/miniredis/v2"
)

func main() {
	addr := flag.String("addr", ":6379", "Address for the Redis-compatible ministore")
	flag.Parse()

	envAddr := os.Getenv("MINISTORE_PORT")
	if envAddr != "" {
		*addr = ":" + envAddr
	}

	s := miniredis.NewMiniRedis()
	if err := s.StartAddr(*addr); err != nil {
		log.Fatalf("ministore: failed to start on %s: %v", *addr, err)
	}

	log.Printf("ministore: Redis-compatible server listening on redis://%s", s.Addr())
	log.Printf("ministore: use REDIS_URL=redis://%s", s.Addr())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("ministore: shutting down...")
	s.Close()
}
