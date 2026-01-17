import requests
import json


class ScopionClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url

    def ingest_event(self, level, service, name, trace_id=None):
        """Ingest a new event."""
        data = {"level": level, "service": service, "name": name}
        if trace_id:
            data["trace_id"] = trace_id
        response = requests.post(f"{self.base_url}/ingest", json=data)
        response.raise_for_status()

    def get_events(self, limit=100):
        """Fetch recent events."""
        response = requests.get(f"{self.base_url}/api/events", params={"limit": limit})
        response.raise_for_status()
        return response.json()

    def subscribe_live(self):
        """Subscribe to live events via SSE."""
        response = requests.get(
            f"{self.base_url}/api/live",
            stream=True,
            headers={"Accept": "text/event-stream"},
        )
        response.raise_for_status()
        for line in response.iter_lines():
            if line.startswith(b"data: "):
                yield json.loads(line[6:].decode("utf-8"))
