# Scopion Go Client

A Go client library for interacting with the Scopion observability service.

## Installation

```bash
go get github.com/xonoxc/scopion/clients/go
```

## Usage

```go
package main

import (
    "log"
    "github.com/xonoxc/scopion/clients/go"
)

func main() {
    client := client.NewClient("http://localhost:8080")

    // Ingest an event
    err := client.IngestEvent("info", "service", "event", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Get events
    events, err := client.GetEvents(10)
    if err != nil {
        log.Fatal(err)
    }
    log.Println(events)

    // Subscribe to live events
    ch, err := client.SubscribeLive()
    if err != nil {
        log.Fatal(err)
    }
    for event := range ch {
        log.Println(event)
    }
}
```

## API

- `NewClient(baseURL string) *Client`: Create a new client.
- `IngestEvent(level, service, name string, traceID *string) error`: Send an event.
- `GetEvents(limit int) ([]Event, error)`: Retrieve recent events.
- `SubscribeLive() (<-chan Event, error)`: Stream live events.

See [Scopion](https://github.com/xonoxc/scopion) for more details.