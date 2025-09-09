package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// Span is a finished server span.
type Span struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	Name         string            `json:"name"`
	Start        time.Time         `json:"start"`
	End          time.Time         `json:"end"`
	DurationMS   float64           `json:"duration_ms"`
	Attrs        map[string]string `json:"attrs,omitempty"`
	Status       string            `json:"status,omitempty"`      // "ok" | "error"
	StatusCode   int               `json:"status_code,omitempty"` // HTTP status
}

// Exporter is where spans are written.
type Exporter interface {
	Export(s Span)
}

// StdoutExporter writes one-line JSON per span to stdout.
type StdoutExporter struct{}

func (StdoutExporter) Export(s Span) {
	b, _ := json.Marshal(s)
	_, _ = os.Stdout.WriteString(string(b) + "\n")
}

// global exporter (can be replaced by the app).
var exporter Exporter = StdoutExporter{}

// SetExporter sets the global span exporter.
func SetExporter(e Exporter) {
	exporter = e
}

// NewTraceID returns a 16-byte (32 hex chars) trace ID.
func NewTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// NewSpanID returns an 8-byte (16 hex chars) span ID.
func NewSpanID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ParseTraceParent parses "00-<traceid>-<spanid>-<flags>" (minimal).
// Returns traceID, parentSpanID, ok.
func ParseTraceParent(v string) (string, string, bool) {
	parts := strings.Split(v, "-")
	if len(parts) < 3 {
		return "", "", false
	}
	traceID := strings.ToLower(parts[1])
	spanID := strings.ToLower(parts[2])
	if len(traceID) != 32 || len(spanID) != 16 {
		return "", "", false
	}
	return traceID, spanID, true
}

// MakeTraceParent builds a traceparent header with sampled flag.
func MakeTraceParent(traceID, spanID string) string {
	return "00-" + strings.ToLower(traceID) + "-" + strings.ToLower(spanID) + "-01"
}

// StatusString maps HTTP status code to "ok" or "error".
func StatusString(code int) string {
	if code >= 500 {
		return "error"
	}
	return "ok"
}

// NewServerSpan builds a Span and exports it.
func NewServerSpan(
	start, end time.Time,
	name, traceID, parentSpanID, spanID string,
	attrs map[string]string,
	statusCode int,
) {
	if attrs == nil {
		attrs = map[string]string{}
	}
	exporter.Export(Span{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Name:         name,
		Start:        start.UTC(),
		End:          end.UTC(),
		DurationMS:   float64(end.Sub(start)) / float64(time.Millisecond),
		Attrs:        attrs,
		Status:       StatusString(statusCode),
		StatusCode:   statusCode,
	})
}

// Str converts basic values to string; useful for attributes.
func Str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
