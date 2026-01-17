# Scopion Python Client

A Python client library for the Scopion observability service.

## Installation

```bash
pip install scopion-client
```

## Usage

```python
from scopion_client import ScopionClient

client = ScopionClient("http://localhost:8080")

# Ingest an event
client.ingest_event("info", "service", "event", trace_id="123")

# Get events
events = client.get_events(10)
print(events)

# Subscribe to live events
for event in client.subscribe_live():
    print(event)
```

## API

- `ScopionClient(base_url)`: Initialize client.
- `ingest_event(level, service, name, trace_id=None)`: Send event.
- `get_events(limit=100)`: Fetch events.
- `subscribe_live()`: Generator for live events.

See [Scopion](https://github.com/xonoxc/scopion) for more.
