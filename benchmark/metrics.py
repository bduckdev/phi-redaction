"""
Metrics calculation for PHI detection benchmark.

Supports exact and overlap-based entity matching.
"""

from dataclasses import dataclass, field
from enum import Enum


class MatchMode(Enum):
    """Entity matching modes."""

    EXACT = "exact"  # Start/end must match exactly
    OVERLAP = "overlap"  # Any overlap counts as match
    TYPE_EXACT = "type_exact"  # Exact match + type must match


@dataclass
class Metrics:
    """Precision, recall, F1 and supporting counts."""

    tp: int = 0  # True positives
    fp: int = 0  # False positives
    fn: int = 0  # False negatives

    @property
    def precision(self) -> float:
        """Precision: tp / (tp + fp)"""
        if self.tp + self.fp == 0:
            return 0.0
        return self.tp / (self.tp + self.fp)

    @property
    def recall(self) -> float:
        """Recall: tp / (tp + fn)"""
        if self.tp + self.fn == 0:
            return 0.0
        return self.tp / (self.tp + self.fn)

    @property
    def f1(self) -> float:
        """F1: harmonic mean of precision and recall"""
        if self.precision + self.recall == 0:
            return 0.0
        return 2 * self.precision * self.recall / (self.precision + self.recall)

    def to_dict(self) -> dict:
        return {
            "tp": self.tp,
            "fp": self.fp,
            "fn": self.fn,
            "precision": round(self.precision, 4),
            "recall": round(self.recall, 4),
            "f1": round(self.f1, 4),
        }


@dataclass
class AggregateMetrics:
    """Aggregated metrics across multiple samples."""

    overall: Metrics = field(default_factory=Metrics)
    by_type: dict[str, Metrics] = field(default_factory=dict)
    latency_ms: list[float] = field(default_factory=list)

    @property
    def avg_latency_ms(self) -> float:
        if not self.latency_ms:
            return 0.0
        return sum(self.latency_ms) / len(self.latency_ms)

    @property
    def p50_latency_ms(self) -> float:
        if not self.latency_ms:
            return 0.0
        sorted_latencies = sorted(self.latency_ms)
        idx = len(sorted_latencies) // 2
        return sorted_latencies[idx]

    @property
    def p95_latency_ms(self) -> float:
        if not self.latency_ms:
            return 0.0
        sorted_latencies = sorted(self.latency_ms)
        idx = int(len(sorted_latencies) * 0.95)
        return sorted_latencies[min(idx, len(sorted_latencies) - 1)]

    @property
    def throughput(self) -> float:
        """Samples per second based on average latency."""
        if self.avg_latency_ms == 0:
            return 0.0
        return 1000.0 / self.avg_latency_ms

    def to_dict(self) -> dict:
        return {
            "overall": self.overall.to_dict(),
            "by_type": {k: v.to_dict() for k, v in self.by_type.items()},
            "latency_ms": {
                "mean": round(self.avg_latency_ms, 2),
                "p50": round(self.p50_latency_ms, 2),
                "p95": round(self.p95_latency_ms, 2),
            },
            "throughput_per_sec": round(self.throughput, 1),
        }


def entities_overlap(a: dict, b: dict) -> bool:
    """Check if two entities overlap in position."""
    return a["start"] < b["end"] and a["end"] > b["start"]


def entities_match(pred: dict, gt: dict, mode: MatchMode) -> bool:
    """Check if predicted entity matches ground truth entity."""
    if mode == MatchMode.EXACT:
        return pred["start"] == gt["start"] and pred["end"] == gt["end"]
    elif mode == MatchMode.OVERLAP:
        return entities_overlap(pred, gt)
    elif mode == MatchMode.TYPE_EXACT:
        return (
            pred["start"] == gt["start"]
            and pred["end"] == gt["end"]
            and pred["type"] == gt["type"]
        )
    return False


def calculate_metrics(
    predicted: list[dict],
    ground_truth: list[dict],
    mode: MatchMode = MatchMode.OVERLAP,
) -> Metrics:
    """
    Calculate precision, recall, F1 for a single sample.

    Args:
        predicted: List of predicted entities
        ground_truth: List of ground truth entities
        mode: Matching mode (exact, overlap, type_exact)

    Returns:
        Metrics object with tp, fp, fn, precision, recall, f1
    """
    metrics = Metrics()
    matched_gt_indices: set[int] = set()

    # Match predictions to ground truth
    for pred in predicted:
        matched = False
        for i, gt in enumerate(ground_truth):
            if i in matched_gt_indices:
                continue
            if entities_match(pred, gt, mode):
                metrics.tp += 1
                matched_gt_indices.add(i)
                matched = True
                break
        if not matched:
            metrics.fp += 1

    # Count false negatives (unmatched ground truth)
    metrics.fn = len(ground_truth) - len(matched_gt_indices)

    return metrics


def calculate_metrics_by_type(
    predicted: list[dict],
    ground_truth: list[dict],
    mode: MatchMode = MatchMode.OVERLAP,
) -> dict[str, Metrics]:
    """
    Calculate metrics broken down by entity type.

    Returns dict mapping entity type to Metrics.
    """
    # Get all types
    all_types = set(e["type"] for e in ground_truth) | set(e["type"] for e in predicted)

    result = {}
    for etype in all_types:
        pred_filtered = [e for e in predicted if e["type"] == etype]
        gt_filtered = [e for e in ground_truth if e["type"] == etype]
        result[etype] = calculate_metrics(pred_filtered, gt_filtered, mode)

    return result


def aggregate_metrics(
    all_metrics: list[Metrics],
    by_type_metrics: list[dict[str, Metrics]],
    latencies: list[float],
) -> AggregateMetrics:
    """
    Aggregate metrics across multiple samples.

    Args:
        all_metrics: List of Metrics from each sample
        by_type_metrics: List of by-type metrics dicts from each sample
        latencies: List of latencies in milliseconds

    Returns:
        AggregateMetrics with overall and by-type metrics
    """
    agg = AggregateMetrics()
    agg.latency_ms = latencies

    # Sum up overall metrics
    for m in all_metrics:
        agg.overall.tp += m.tp
        agg.overall.fp += m.fp
        agg.overall.fn += m.fn

    # Sum up by-type metrics
    for by_type in by_type_metrics:
        for etype, m in by_type.items():
            if etype not in agg.by_type:
                agg.by_type[etype] = Metrics()
            agg.by_type[etype].tp += m.tp
            agg.by_type[etype].fp += m.fp
            agg.by_type[etype].fn += m.fn

    return agg
