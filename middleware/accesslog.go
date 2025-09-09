package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aminofox/zentrox"
)

// AccessLogConfig configures access logging.
type AccessLogConfig struct {
	Format           string       // "text" or "json"
	TimeFormat       string       // used when Format == "text"
	IncludeRequestID bool         // add request_id if present
	LogFunc          func(string) // override sink; default: stdout
	ContextKeyRID    string       // where request id is stored; default: "request_id"
}

func DefaultAccessLog() AccessLogConfig {
	return AccessLogConfig{
		Format:           "text",
		TimeFormat:       time.RFC3339,
		IncludeRequestID: true,
		ContextKeyRID:    zentrox.RequestID,
	}
}

func AccessLog(cfg AccessLogConfig) zentrox.Handler {
	if cfg.Format != "json" {
		cfg.Format = "text"
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	if cfg.LogFunc == nil {
		cfg.LogFunc = func(s string) { _, _ = os.Stdout.WriteString(s + "\n") }
	}
	if cfg.ContextKeyRID == "" {
		cfg.ContextKeyRID = zentrox.RequestID
	}

	return func(c *zentrox.Context) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: c.Writer}
		c.Writer = sw

		c.Forward()

		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		lat := time.Since(start)

		method := c.Request.Method
		path := c.Request.URL.Path
		ua := c.Request.UserAgent()
		ip := clientIP(c.Request)

		rid := ""
		if cfg.IncludeRequestID {
			if v, ok := c.Get(cfg.ContextKeyRID); ok {
				if s, _ := v.(string); s != "" {
					rid = s
				}
			}
		}

		ver := ""
		if v, ok := c.Get(zentrox.AppVersion); ok {
			if s, _ := v.(string); s != "" {
				ver = s
			}
		}

		if cfg.Format == "json" {
			rec := map[string]any{
				"ts":      time.Now().Format(time.RFC3339Nano),
				"method":  method,
				"path":    path,
				"status":  sw.status,
				"bytes":   sw.bytes,
				"latency": float64(lat) / float64(time.Millisecond),
				"ip":      ip,
				"ua":      ua,
			}
			if rid != "" {
				rec[zentrox.RequestID] = rid
			}
			// NEW: include version field if present
			if ver != "" {
				rec["version"] = ver
			}
			b, _ := json.Marshal(rec)
			cfg.LogFunc(string(b))
			return
		} else {
			ts := time.Now().Format(cfg.TimeFormat)
			if rid != "" {
				// with request id
				if ver != "" {
					cfg.LogFunc(fmt.Sprintf("%s | %s %s | %d %dB | %v | ip=%s | rid=%s | ver=%s | ua=%q",
						ts, method, path, sw.status, sw.bytes, lat, ip, rid, ver, ua))
				} else {
					cfg.LogFunc(fmt.Sprintf("%s | %s %s | %d %dB | %v | ip=%s | rid=%s | ua=%q",
						ts, method, path, sw.status, sw.bytes, lat, ip, rid, ua))
				}
				return
			}

			// without request id
			if ver != "" {
				cfg.LogFunc(fmt.Sprintf("%s | %s %s | %d %dB | %v | ip=%s | ver=%s | ua=%q",
					ts, method, path, sw.status, sw.bytes, lat, ip, ver, ua))
			} else {
				cfg.LogFunc(fmt.Sprintf("%s | %s %s | %d %dB | %v | ip=%s | ua=%q",
					ts, method, path, sw.status, sw.bytes, lat, ip, ua))
			}
		}

	}
}
