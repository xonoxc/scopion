interface Event {
  id: string;
  timestamp: string;
  level: string;
  service: string;
  name: string;
  trace_id?: string;
}

class ScopionClient {
  constructor(private baseUrl: string = "http://localhost:8080") {}

  async ingestEvent(level: string, service: string, name: string, traceId?: string): Promise<void> {
    const data: any = { level, service, name };
    if (traceId) data.trace_id = traceId;
    const response = await fetch(`${this.baseUrl}/ingest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    if (!response.ok) throw new Error(`Failed to ingest: ${response.status}`);
  }

  async getEvents(limit: number = 100): Promise<Event[]> {
    const response = await fetch(`${this.baseUrl}/api/events?limit=${limit}`);
    if (!response.ok) throw new Error(`Failed to get events: ${response.status}`);
    return response.json();
  }

  subscribeLive(onEvent: (event: Event) => void): EventSource {
    const eventSource = new EventSource(`${this.baseUrl}/api/live`);
    eventSource.onmessage = (e) => {
      try {
        const event: Event = JSON.parse(e.data);
        onEvent(event);
      } catch (err) {
        console.error('Failed to parse event:', err);
      }
    };
    return eventSource;
  }
}

export { ScopionClient, Event };