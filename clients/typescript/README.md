# Scopion TypeScript Client

A TypeScript client for the Scopion observability service.

## Installation

```bash
npm install scopion-client
```

## Usage

```typescript
import { ScopionClient } from 'scopion-client';

const client = new ScopionClient('http://localhost:8080');

// Ingest event
await client.ingestEvent('info', 'service', 'event', 'trace123');

// Get events
const events = await client.getEvents(10);
console.log(events);

// Subscribe to live events
const source = client.subscribeLive(event => console.log(event));
```

## API

- `new ScopionClient(baseUrl)`: Create client.
- `ingestEvent(level, service, name, traceId?)`: Send event.
- `getEvents(limit?)`: Get events.
- `subscribeLive(callback)`: Subscribe to live events.

See [Scopion](https://github.com/xonoxc/scopion) for more.
