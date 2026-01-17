import { ScopionClient } from "./scopion";

describe("ScopionClient", () => {
  let client: ScopionClient;

  beforeEach(() => {
    client = new ScopionClient("http://test");
  });

  test("ingestEvent", async () => {
    global.fetch = jest.fn(() => Promise.resolve({ ok: true } as Response));
    await client.ingestEvent("info", "test", "event");
    expect(fetch).toHaveBeenCalledWith("http://test/ingest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ level: "info", service: "test", name: "event" }),
    });
  });

  test("ingestEvent with trace", async () => {
    global.fetch = jest.fn(() => Promise.resolve({ ok: true } as Response));
    await client.ingestEvent("error", "api", "timeout", "trace123");
    expect(fetch).toHaveBeenCalledWith("http://test/ingest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        level: "error",
        service: "api",
        name: "timeout",
        trace_id: "trace123",
      }),
    });
  });

  test("getEvents", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve([{ id: "1", level: "info" }]),
      } as Response),
    );
    const events = await client.getEvents(50);
    expect(events).toEqual([{ id: "1", level: "info" }]);
    expect(fetch).toHaveBeenCalledWith("http://test/api/events?limit=50");
  });

  test("subscribeLive", () => {
    const mockEventSource = {
      onmessage: jest.fn(),
      close: jest.fn(),
    };
    global.EventSource = jest.fn(() => mockEventSource) as any;
    const onEvent = jest.fn();
    const _ = client.subscribeLive(onEvent);
    expect(EventSource).toHaveBeenCalledWith("http://test/api/live");
    // Simulate message
    mockEventSource.onmessage({ data: '{"id": "1"}' } as MessageEvent);
    expect(onEvent).toHaveBeenCalledWith({ id: "1" });
  });
});

