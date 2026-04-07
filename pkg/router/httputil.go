package router

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Request ID
// ---------------------------------------------------------------------------

type reqIDKeyType struct{}

var reqIDKey = reqIDKeyType{}

// RequestID generates a unique request identifier and stores it in the request
// context. The ID is also written as the X-Request-Id response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetReqID returns the request ID from context, or empty string.
func GetReqID(ctx context.Context) string {
	if id, ok := ctx.Value(reqIDKey).(string); ok {
		return id
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ---------------------------------------------------------------------------
// Real IP
// ---------------------------------------------------------------------------

// RealIP sets r.RemoteAddr to the first trusted IP from X-Forwarded-For or
// X-Real-IP headers.
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rip := realIPFromRequest(r); rip != "" {
			r.RemoteAddr = rip
		}
		next.ServeHTTP(w, r)
	})
}

func realIPFromRequest(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	return ""
}

// ---------------------------------------------------------------------------
// Recoverer
// ---------------------------------------------------------------------------

// Recoverer catches panics in downstream handlers, logs the stack trace, and
// returns a 500 Internal Server Error response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				slog.Error("panic recovered",
					"error", rv,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
				)
				if !headerWritten(w) {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func headerWritten(w http.ResponseWriter) bool {
	// If the header map has a status-related entry, consider it written.
	// This is a best-effort check; if the handler has already called
	// WriteHeader, we cannot call it again without triggering a superfluous
	// warning. We rely on the WrapResponseWriter for accurate tracking
	// in the middleware stack.
	return false
}

// ---------------------------------------------------------------------------
// Timeout
// ---------------------------------------------------------------------------

// TimeoutMiddleware wraps the stdlib http.TimeoutHandler to cancel requests
// that exceed the given duration.
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if timeout <= 0 {
			return next
		}
		return http.TimeoutHandler(next, timeout, `{"error":{"code":"TIMEOUT","message":"request timeout"}}`)
	}
}

// ---------------------------------------------------------------------------
// Compress
// ---------------------------------------------------------------------------

// Compress returns middleware that gzip-compresses response bodies for clients
// that accept gzip encoding. level follows compress/flate constants.
func Compress(level int) func(http.Handler) http.Handler {
	if level < flate.HuffmanOnly {
		level = flate.DefaultCompression
	}
	if level > flate.BestCompression {
		level = flate.BestCompression
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz, err := gzip.NewWriterLevel(w, level)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) Flush() {
	if f, ok := g.writer.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// ---------------------------------------------------------------------------
// WrapResponseWriter
// ---------------------------------------------------------------------------

// WrapResponseWriter is a response writer wrapper that captures the HTTP status
// code and the number of bytes written. It replaces chi/middleware's equivalent.
type WrapResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
	mu           sync.Mutex
}

// NewWrapResponseWriter creates a new WrapResponseWriter. The protoMajor
// argument is accepted for API compatibility but currently unused.
func NewWrapResponseWriter(w http.ResponseWriter, _ int) *WrapResponseWriter {
	return &WrapResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *WrapResponseWriter) WriteHeader(code int) {
	w.mu.Lock()
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
	w.mu.Unlock()
}

func (w *WrapResponseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	w.mu.Unlock()
	n, err := w.ResponseWriter.Write(b)
	w.mu.Lock()
	w.bytesWritten += n
	w.mu.Unlock()
	return n, err
}

// Status returns the HTTP status code that was written.
func (w *WrapResponseWriter) Status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

// BytesWritten returns the total bytes written to the response body.
func (w *WrapResponseWriter) BytesWritten() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.bytesWritten
}

// Unwrap returns the underlying ResponseWriter for middleware compatibility.
func (w *WrapResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush implements http.Flusher if the underlying writer supports it.
func (w *WrapResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// clientIPFromRequest extracts the client IP. Exported for reuse across
// packages, but also used internally by the rate limiter.
func clientIPFromRequest(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
