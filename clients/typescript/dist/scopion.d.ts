interface Event {
    id: string;
    timestamp: string;
    level: string;
    service: string;
    name: string;
    trace_id?: string;
}
declare class ScopionClient {
    private baseUrl;
    constructor(baseUrl?: string);
    ingestEvent(level: string, service: string, name: string, traceId?: string): Promise<void>;
    getEvents(limit?: number): Promise<Event[]>;
    subscribeLive(onEvent: (event: Event) => void): EventSource;
}
export { ScopionClient, Event };
