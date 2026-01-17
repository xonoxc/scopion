import unittest
from unittest.mock import patch, MagicMock
import sys
import os

sys.path.insert(0, os.path.dirname(__file__))
from scopion_client import ScopionClient


class TestScopionClient(unittest.TestCase):
    def setUp(self):
        self.client = ScopionClient("http://test")

    @patch("scopion_client.requests.post")
    def test_ingest_event(self, mock_post):
        mock_post.return_value.status_code = 202
        self.client.ingest_event("info", "test", "event")
        mock_post.assert_called_once_with(
            "http://test/ingest",
            json={"level": "info", "service": "test", "name": "event"},
        )

    @patch("scopion_client.requests.post")
    def test_ingest_event_with_trace(self, mock_post):
        mock_post.return_value.status_code = 202
        self.client.ingest_event("error", "api", "timeout", "trace123")
        mock_post.assert_called_once_with(
            "http://test/ingest",
            json={
                "level": "error",
                "service": "api",
                "name": "timeout",
                "trace_id": "trace123",
            },
        )

    @patch("scopion_client.requests.get")
    def test_get_events(self, mock_get):
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = [{"id": "1", "level": "info"}]
        events = self.client.get_events(50)
        self.assertEqual(events, [{"id": "1", "level": "info"}])
        mock_get.assert_called_once_with("http://test/api/events", params={"limit": 50})

    @patch("scopion_client.requests.get")
    def test_subscribe_live(self, mock_get):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.iter_lines.return_value = [
            b'data: {"id": "1", "level": "info"}',
            b"",
            b'data: {"id": "2"}',
        ]
        mock_get.return_value = mock_response
        gen = self.client.subscribe_live()
        self.assertEqual(next(gen), {"id": "1", "level": "info"})
        self.assertEqual(next(gen), {"id": "2"})


if __name__ == "__main__":
    unittest.main()
