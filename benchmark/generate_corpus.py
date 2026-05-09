#!/usr/bin/env python3
"""
Synthetic PHI corpus generator for benchmarking.

Generates labeled samples with known PHI entities at known positions.
"""

import argparse
import json
import random
import re
from dataclasses import dataclass
from pathlib import Path

from faker import Faker

fake = Faker()


@dataclass
class Entity:
    type: str
    text: str
    start: int
    end: int

    def to_dict(self) -> dict:
        return {
            "type": self.type,
            "text": self.text,
            "start": self.start,
            "end": self.end,
        }


# Templates with placeholders for PHI
# Format: {TYPE} will be replaced with generated data
TEMPLATES = [
    "Please contact {NAME} at {EMAIL} regarding the appointment.",
    "Patient {NAME} (SSN: {SSN}) called from {PHONE}.",
    "Dr. {NAME} can be reached at {EMAIL} or {PHONE}.",
    "{NAME} reported symptoms. Contact: {EMAIL}",
    "The patient, {NAME}, provided SSN {SSN} for verification.",
    "Referral from {NAME} ({EMAIL}) for new patient intake.",
    "Call {NAME} back at {PHONE} to confirm lab results.",
    "Insurance verification needed for {NAME}, SSN ending in {SSN}.",
    "Message for {NAME}: Please call {PHONE} at your earliest convenience.",
    "{NAME} requested records be sent to {EMAIL}.",
    "Appointment reminder for {NAME} - contact {PHONE} if unable to attend.",
    "Patient {NAME} (contact: {EMAIL}, {PHONE}) scheduled for follow-up.",
    "Prescription ready for {NAME}. Verify identity with SSN {SSN}.",
    "Transfer request from Dr. {NAME} at {EMAIL}.",
    "Emergency contact for {NAME}: {PHONE}.",
    "Lab results for {NAME} sent to {EMAIL}.",
    "{NAME} confirmed address and SSN {SSN} on file.",
    "Billing inquiry from {NAME}, reachable at {PHONE} or {EMAIL}.",
    "Prior authorization needed. Patient: {NAME}, SSN: {SSN}.",
    "Specialist referral: {NAME} requests callback at {PHONE}.",
]

# Additional context sentences (no PHI) to mix in
CONTEXT_SENTENCES = [
    "The appointment is scheduled for next Tuesday.",
    "Please bring your insurance card.",
    "Lab results will be ready in 3-5 business days.",
    "The office is open Monday through Friday.",
    "A copay may be required at the time of visit.",
    "Please arrive 15 minutes early to complete paperwork.",
    "Fasting is required before the blood draw.",
    "The referral has been submitted to your insurance.",
]


def generate_name() -> str:
    """Generate a realistic name."""
    choice = random.choice(["full", "full", "first_last", "first_last", "titled"])
    if choice == "full":
        return fake.name()
    elif choice == "first_last":
        return f"{fake.first_name()} {fake.last_name()}"
    else:
        return f"Dr. {fake.last_name()}"


def generate_email() -> str:
    """Generate a realistic email."""
    return fake.email()


def generate_phone() -> str:
    """Generate a phone number in various formats."""
    formats = [
        "###-###-####",
        "(###) ###-####",
        "### ### ####",
        "###.###.####",
        "1-###-###-####",
    ]
    return fake.numerify(random.choice(formats))


def generate_ssn() -> str:
    """
    Generate a valid SSN in XXX-XX-XXXX format.

    Avoids invalid patterns that Presidio rejects:
    - Area numbers 000, 666, 900-999
    - Group number 00
    - Serial number 0000
    - Sequential patterns (123456789, etc.)
    - Famous leaked SSN (078-05-1120)
    """
    while True:
        # Generate area (001-665, 667-899)
        area = random.randint(1, 899)
        if area in (0, 666) or area >= 900:
            continue

        # Generate group (01-99)
        group = random.randint(1, 99)

        # Generate serial (0001-9999)
        serial = random.randint(1, 9999)

        ssn = f"{area:03d}-{group:02d}-{serial:04d}"
        digits = f"{area:03d}{group:02d}{serial:04d}"

        # Skip invalid patterns
        if all(c == digits[0] for c in digits):
            continue
        if digits in ("123456789", "987654321", "078051120"):
            continue

        return ssn


GENERATORS = {
    "NAME": generate_name,
    "EMAIL": generate_email,
    "PHONE": generate_phone,
    "SSN": generate_ssn,
}


def generate_sample(sample_id: str) -> dict:
    """Generate a single labeled sample."""
    template = random.choice(TEMPLATES)

    # Find all placeholders
    placeholders = re.findall(r"\{(\w+)\}", template)

    # Generate values and track positions
    text = template
    entities = []

    # Process each placeholder
    for placeholder in placeholders:
        if placeholder not in GENERATORS:
            continue

        # Generate value
        value = GENERATORS[placeholder]()

        # Find position of placeholder in current text
        pattern = "{" + placeholder + "}"
        match = re.search(re.escape(pattern), text)
        if not match:
            continue

        start = match.start()
        end = start + len(value)

        # Replace placeholder with value
        text = text[: match.start()] + value + text[match.end() :]

        # Record entity
        entities.append(
            Entity(
                type=placeholder,
                text=value,
                start=start,
                end=end,
            )
        )

    # Optionally add context (no PHI) before or after
    if random.random() < 0.3:
        context = random.choice(CONTEXT_SENTENCES)
        if random.random() < 0.5:
            # Prepend context
            offset = len(context) + 1  # +1 for space
            text = context + " " + text
            for e in entities:
                e.start += offset
                e.end += offset
        else:
            # Append context
            text = text + " " + context

    return {
        "id": sample_id,
        "text": text,
        "entities": [e.to_dict() for e in entities],
    }


def generate_corpus(count: int, output_path: Path) -> None:
    """Generate corpus and write to JSONL file."""
    output_path.parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, "w") as f:
        for i in range(count):
            sample = generate_sample(f"sample_{i:05d}")
            f.write(json.dumps(sample) + "\n")

    print(f"Generated {count} samples to {output_path}")

    # Print sample stats
    entity_counts = {"NAME": 0, "EMAIL": 0, "PHONE": 0, "SSN": 0}
    with open(output_path) as f:
        for line in f:
            sample = json.loads(line)
            for entity in sample["entities"]:
                entity_counts[entity["type"]] = (
                    entity_counts.get(entity["type"], 0) + 1
                )

    print("Entity distribution:")
    for etype, count in sorted(entity_counts.items()):
        print(f"  {etype}: {count}")


def main():
    parser = argparse.ArgumentParser(description="Generate synthetic PHI corpus")
    parser.add_argument(
        "--count", type=int, default=1000, help="Number of samples to generate"
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("corpus/labeled_samples.jsonl"),
        help="Output file path",
    )
    parser.add_argument("--seed", type=int, default=42, help="Random seed")

    args = parser.parse_args()

    # Set seeds for reproducibility
    random.seed(args.seed)
    Faker.seed(args.seed)

    generate_corpus(args.count, args.output)


if __name__ == "__main__":
    main()
