"""
Client for the phi-redactor Go API.
"""

import requests


class PhiRedactor:
    """Client for the phi-redactor API."""

    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()

    def extract(self, text: str) -> list[dict]:
        """
        Extract PHI entities from text.

        Returns list of entities with type, text, start, end.
        """
        resp = self.session.post(
            f"{self.base_url}/redact",
            json={"text": text, "options": {"include_detail": True}},
            timeout=30,
        )
        resp.raise_for_status()
        data = resp.json()

        return [
            {
                "type": f["type"],
                "text": f["text"],
                "start": f["start"],
                "end": f["end"],
            }
            for f in data.get("findings", [])
        ]

    def health_check(self) -> bool:
        """Check if the API is healthy."""
        try:
            resp = self.session.get(f"{self.base_url}/healthz", timeout=5)
            return resp.status_code == 200
        except requests.RequestException:
            return False
