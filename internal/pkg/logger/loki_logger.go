package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type LokiHandler struct {
	url    string
	labels map[string]string
	client *http.Client
	mu     sync.Mutex
	level  slog.Level
}

func NewLokiHandler(url string, labels map[string]string) *LokiHandler {
	return &LokiHandler{
		url:    url + "/loki/api/v1/push",
		labels: labels,
		client: &http.Client{Timeout: 5 * time.Second},
		level:  slog.LevelInfo,
	}
}

func (h *LokiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *LokiHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": h.labels,
				"values": [][]string{
					{
						fmt.Sprintf("%d", r.Time.UnixNano()),
						h.formatRecord(r),
					},
				},
			},
		},
	}

	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("loki responded with status %d", resp.StatusCode)
	}

	return nil
}

func (h *LokiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newLabels := make(map[string]string)
	for k, v := range h.labels {
		newLabels[k] = v
	}

	for _, attr := range attrs {
		newLabels[attr.Key] = fmt.Sprint(attr.Value.Any())
	}

	return &LokiHandler{
		url:    h.url,
		labels: newLabels,
		client: h.client,
		level:  h.level,
	}
}

func (h *LokiHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *LokiHandler) formatRecord(r slog.Record) string {
	logEntry := map[string]interface{}{
		"level":   r.Level.String(),
		"message": r.Message,
		"time":    r.Time.Format(time.RFC3339),
	}

	r.Attrs(func(attr slog.Attr) bool {
		logEntry[attr.Key] = attr.Value.Any()
		return true
	})

	jsonData, _ := json.Marshal(logEntry)
	return string(jsonData)
}
