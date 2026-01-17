class ScopionClient {
    constructor(baseUrl = "http://localhost:8080") {
        this.baseUrl = baseUrl;
    }
    async ingestEvent(level, service, name, traceId) {
        const data = { level, service, name };
        if (traceId)
            data.trace_id = traceId;
        const response = await fetch(`${this.baseUrl}/ingest`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!response.ok)
            throw new Error(`Failed to ingest: ${response.status}`);
    }
    async getEvents(limit = 100) {
        const response = await fetch(`${this.baseUrl}/api/events?limit=${limit}`);
        if (!response.ok)
            throw new Error(`Failed to get events: ${response.status}`);
        return response.json();
    }
    subscribeLive(onEvent) {
        const eventSource = new EventSource(`${this.baseUrl}/api/live`);
        eventSource.onmessage = (e) => {
            try {
                const event = JSON.parse(e.data);
                onEvent(event);
            }
            catch (err) {
                console.error('Failed to parse event:', err);
            }
        };
        return eventSource;
    }
}
export { ScopionClient };
