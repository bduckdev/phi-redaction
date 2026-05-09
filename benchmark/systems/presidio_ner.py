"""
Presidio-based PHI detection wrapper.
"""

from presidio_analyzer import AnalyzerEngine


class PresidioNER:
    """Presidio analyzer wrapper for PHI detection."""

    # Map Presidio entity types to our standard types
    TYPE_MAPPING = {
        "PERSON": "NAME",
        "EMAIL_ADDRESS": "EMAIL",
        "PHONE_NUMBER": "PHONE",
        "US_SSN": "SSN",
    }

    # Entity types we care about
    ENTITIES = ["PERSON", "EMAIL_ADDRESS", "PHONE_NUMBER", "US_SSN"]

    def __init__(self):
        self.analyzer = AnalyzerEngine()

    def extract(self, text: str, score_threshold: float = 0.0) -> list[dict]:
        """
        Extract PHI entities from text.

        Returns list of entities with type, text, start, end, score.

        Args:
            text: Input text to analyze
            score_threshold: Minimum score to include (0.0 = include all)
        """
        results = self.analyzer.analyze(
            text=text,
            entities=self.ENTITIES,
            language="en",
            score_threshold=score_threshold,
        )

        return [
            {
                "type": self._map_type(r.entity_type),
                "text": text[r.start : r.end],
                "start": r.start,
                "end": r.end,
                "score": r.score,
            }
            for r in results
            if r.entity_type in self.TYPE_MAPPING
        ]

    def _map_type(self, entity_type: str) -> str:
        """Map Presidio entity type to our standard type."""
        return self.TYPE_MAPPING.get(entity_type, entity_type)
