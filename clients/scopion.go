package scopion

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

type Event struct {
	ID        string  `json:"id"`
	Timestamp string  `json:"timestamp"`
	Level     string  `json:"level"`
	Service   string  `json:"service"`
	Name      string  `json:"name"`
	TraceID   *string `json:"trace_id,omitempty"`
}

func (c *Client) IngestEvent(level, service, name string, traceID *string) error {
	data := map[string]any{
		"level":   level,
		"service": service,
		"name":    name,
	}
	if traceID != nil {
		data["trace_id"] = *traceID
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.BaseURL+"/ingest", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
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

func (c *Client) SubscribeLive() (<-chan Event, error) {
	ch := make(chan Event)
	go func() {
		defer close(ch)
		resp, err := http.Get(c.BaseURL + "/api/live")
		if err != nil {
			return
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if after, ok := strings.CutPrefix(line, "data: "); ok {
				data := after
				var event Event
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					ch <- event
				}
			}
		}
	}()
	return ch, nil
}
