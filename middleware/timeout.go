package middleware

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aminofox/zentrox"
)

type TimeoutOptions struct {
	Duration time.Duration
}

func Timeout(opt TimeoutOptions) zentrox.Handler {
	if opt.Duration <= 0 {
		opt.Duration = 5 * time.Second
	}
	return func(c *zentrox.Context) {
		// create context with timeout for handler
		ctx, cancel := context.WithTimeout(c.Request.Context(), opt.Duration)
		defer cancel()

		// wrap request with ctx
		c.Request = c.Request.WithContext(ctx)

		tw := &timeoutWriter{
			ResponseWriter: c.Writer,
			hdr:            make(http.Header),
		}
		c.Writer = tw

		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Forward()
		}()

		select {
		case <-done:
			// handler done before timeout
			return
		case <-ctx.Done():
			// timed out: send 504 if handler haven't write
			tw.sendTimeout()
			<-done
			return
		}
	}
}

type timeoutWriter struct {
	http.ResponseWriter
	mu          sync.Mutex
	hdr         http.Header
	wroteHeader bool
	timedOut    bool
	status      int
}

func (w *timeoutWriter) Header() http.Header { return w.hdr }

func (w *timeoutWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut || w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = code
	dst := w.ResponseWriter.Header()
	for k, vals := range w.hdr {
		vv := make([]string, len(vals))
		copy(vv, vals)
		dst[k] = vv
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = http.StatusOK
		dst := w.ResponseWriter.Header()
		for k, vals := range w.hdr {
			vv := make([]string, len(vals))
			copy(vv, vals)
			dst[k] = vv
		}
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *timeoutWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *timeoutWriter) sendTimeout() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return
	}
	w.timedOut = true
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = http.StatusGatewayTimeout
		h := w.ResponseWriter.Header()
		if h.Get("Content-Type") == "" {
			h.Set("Content-Type", "text/plain; charset=utf-8")
		}
		w.ResponseWriter.WriteHeader(http.StatusGatewayTimeout)
	}
	_, _ = io.WriteString(w.ResponseWriter, "Gateway Timeout")
}
