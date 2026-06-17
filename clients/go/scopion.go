package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	queue      chan map[string]any
	flushEvery time.Duration
	maxBatch   int
	wg         sync.WaitGroup
	once       sync.Once
	stop       chan struct{}
}

type Event struct {
	ID        string         `json:"id"`
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Service   string         `json:"service"`
	Name      string         `json:"name"`
	TraceID   *string        `json:"trace_id,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

func NewClient(baseURL string) *Client {
	c := &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		queue:      make(chan map[string]any, 4096),
		flushEvery: 500 * time.Millisecond,
		maxBatch:   100,
		stop:       make(chan struct{}),
	}
	c.wg.Add(1)
	go c.bgWorker()
	return c
}

func (c *Client) IngestEvent(level, service, name string, traceID *string, customData map[string]any) error {
	data := map[string]any{
		"level":   level,
		"service": service,
		"name":    name,
	}

	if traceID != nil {
		data["trace_id"] = *traceID
	}

	if customData != nil {
		data["data"] = customData
	}

	select {
	case c.queue <- data:
	default:
		return fmt.Errorf("client queue full, dropping event")
	}
	return nil
}

func (c *Client) bgWorker() {
	defer c.wg.Done()
	batch := make([]map[string]any, 0, c.maxBatch)
	ticker := time.NewTicker(c.flushEvery)
	defer ticker.Stop()

	for {
		select {
		case <-c.stop:
			c.flush(batch)
			return
		case evt := <-c.queue:
			batch = append(batch, evt)
			if len(batch) >= c.maxBatch {
				c.flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				c.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (c *Client) flush(batch []map[string]any) {
	if len(batch) == 0 {
		return
	}

	for _, evt := range batch {
		c.sendWithRetry(evt, 3)
	}
}

func (c *Client) sendWithRetry(data map[string]any, maxRetries int) {
	var lastErr error
	for range maxRetries {
		jsonData, err := json.Marshal(data)
		if err != nil {
			lastErr = err
			break
		}

		resp, err := c.HTTPClient.Post(c.BaseURL+"/ingest", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 202 {
			return
		}

		lastErr = fmt.Errorf("unexpected status: %d", resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	}

	if lastErr != nil {
		log.Printf("failed to send event after retries: %v", lastErr)
	}
}

func (c *Client) Flush() {
	close(c.stop)
	c.wg.Wait()
}

func (c *Client) GetEvents(limit int) ([]Event, error) {
	url := fmt.Sprintf("%s/api/events?limit=%d", c.BaseURL, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var events []Event
	err = json.NewDecoder(resp.Body).Decode(&events)
	return events, err
}
