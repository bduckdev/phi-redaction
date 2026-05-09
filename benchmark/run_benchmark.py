#!/usr/bin/env python3
"""
PHI detection benchmark harness.

Compares phi-redactor (Go API) against Presidio.
"""

import argparse
import json
import sys
import time
from datetime import datetime
from pathlib import Path

from tabulate import tabulate

from metrics import (
    AggregateMetrics,
    MatchMode,
    aggregate_metrics,
    calculate_metrics,
    calculate_metrics_by_type,
)
from systems.phi_redactor import PhiRedactor
from systems.presidio_ner import PresidioNER


def load_corpus(path: Path) -> list[dict]:
    """Load corpus from JSONL file."""
    samples = []
    with open(path) as f:
        for line in f:
            samples.append(json.loads(line))
    return samples


def run_system(system, corpus: list[dict], mode: MatchMode) -> AggregateMetrics:
    """Run a detection system against the corpus and calculate metrics."""
    all_metrics = []
    by_type_metrics = []
    latencies = []

    for sample in corpus:
        text = sample["text"]
        ground_truth = sample["entities"]

        # Time the extraction
        start = time.perf_counter()
        try:
            predicted = system.extract(text)
        except Exception as e:
            print(f"Error processing sample {sample['id']}: {e}", file=sys.stderr)
            predicted = []
        latency_ms = (time.perf_counter() - start) * 1000
        latencies.append(latency_ms)

        # Calculate metrics
        metrics = calculate_metrics(predicted, ground_truth, mode)
        by_type = calculate_metrics_by_type(predicted, ground_truth, mode)

        all_metrics.append(metrics)
        by_type_metrics.append(by_type)

    return aggregate_metrics(all_metrics, by_type_metrics, latencies)


def print_results(results: dict[str, AggregateMetrics]) -> None:
    """Print results as formatted tables."""
    # Overall comparison table
    headers = ["System", "Precision", "Recall", "F1", "Avg Latency", "Throughput"]
    rows = []
    for name, agg in results.items():
        rows.append(
            [
                name,
                f"{agg.overall.precision:.3f}",
                f"{agg.overall.recall:.3f}",
                f"{agg.overall.f1:.3f}",
                f"{agg.avg_latency_ms:.2f}ms",
                f"{agg.throughput:.1f}/s",
            ]
        )

    print("\n=== Overall Results ===")
    print(tabulate(rows, headers=headers, tablefmt="simple"))

    # By-type breakdown for each system
    for name, agg in results.items():
        print(f"\n=== {name} by Entity Type ===")
        type_headers = ["Type", "Precision", "Recall", "F1", "TP", "FP", "FN"]
        type_rows = []
        for etype in sorted(agg.by_type.keys()):
            m = agg.by_type[etype]
            type_rows.append(
                [
                    etype,
                    f"{m.precision:.3f}",
                    f"{m.recall:.3f}",
                    f"{m.f1:.3f}",
                    m.tp,
                    m.fp,
                    m.fn,
                ]
            )
        print(tabulate(type_rows, headers=type_headers, tablefmt="simple"))

    # Latency comparison
    print("\n=== Latency Distribution (ms) ===")
    latency_headers = ["System", "Mean", "P50", "P95"]
    latency_rows = []
    for name, agg in results.items():
        latency_rows.append(
            [
                name,
                f"{agg.avg_latency_ms:.2f}",
                f"{agg.p50_latency_ms:.2f}",
                f"{agg.p95_latency_ms:.2f}",
            ]
        )
    print(tabulate(latency_rows, headers=latency_headers, tablefmt="simple"))


def save_results(
    results: dict[str, AggregateMetrics], corpus_size: int, output_path: Path
) -> None:
    """Save results to JSON file."""
    output = {
        "timestamp": datetime.now().isoformat(),
        "corpus_size": corpus_size,
        "systems": {name: agg.to_dict() for name, agg in results.items()},
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with open(output_path, "w") as f:
        json.dump(output, f, indent=2)

    print(f"\nResults saved to {output_path}")


def main():
    parser = argparse.ArgumentParser(description="Run PHI detection benchmark")
    parser.add_argument(
        "--corpus",
        type=Path,
        default=Path("corpus/labeled_samples.jsonl"),
        help="Path to labeled corpus",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("results/comparison.json"),
        help="Output file for results",
    )
    parser.add_argument(
        "--match-mode",
        choices=["exact", "overlap", "type_exact"],
        default="overlap",
        help="Entity matching mode",
    )
    parser.add_argument(
        "--systems",
        nargs="+",
        choices=["phi-redactor", "presidio"],
        default=["phi-redactor", "presidio"],
        help="Systems to benchmark",
    )
    parser.add_argument(
        "--api-url",
        default="http://localhost:8080",
        help="phi-redactor API URL",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=None,
        help="Limit number of samples (for quick testing)",
    )

    args = parser.parse_args()

    # Parse match mode
    mode = MatchMode(args.match_mode)

    # Load corpus
    print(f"Loading corpus from {args.corpus}...")
    corpus = load_corpus(args.corpus)
    if args.limit:
        corpus = corpus[: args.limit]
    print(f"Loaded {len(corpus)} samples")

    # Initialize systems
    systems = {}

    if "phi-redactor" in args.systems:
        print(f"\nInitializing phi-redactor client ({args.api_url})...")
        phi = PhiRedactor(args.api_url)
        if not phi.health_check():
            print(
                "WARNING: phi-redactor API not responding. Is the server running?",
                file=sys.stderr,
            )
            print("Start with: make run (in the project root)", file=sys.stderr)
            if len(args.systems) == 1:
                sys.exit(1)
        else:
            systems["phi-redactor"] = phi

    if "presidio" in args.systems:
        print("\nInitializing Presidio analyzer...")
        systems["presidio"] = PresidioNER()

    if not systems:
        print("No systems available to benchmark", file=sys.stderr)
        sys.exit(1)

    # Run benchmarks
    results = {}
    for name, system in systems.items():
        print(f"\nRunning {name}...")
        results[name] = run_system(system, corpus, mode)

    # Output results
    print_results(results)
    save_results(results, len(corpus), args.output)


if __name__ == "__main__":
    main()
