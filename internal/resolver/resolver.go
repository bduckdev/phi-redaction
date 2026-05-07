package resolver

import (
	"phi-redactor/internal/candidate"
	"phi-redactor/internal/finding"
)

// Config controls resolver behavior
type Config struct {
	// IncludeHeuristic determines whether low-confidence candidates are emitted
	// When false, only deterministic and candidate confidence findings are returned
	IncludeHeuristic bool
}

// DefaultConfig returns the default resolver configuration
func DefaultConfig() Config {
	return Config{
		IncludeHeuristic: true,
	}
}

// Resolve processes candidates and emits findings with appropriate confidence levels
// It analyzes context signals, detects name pairs, and applies confidence rules
func Resolve(candidates []candidate.Candidate) []finding.Finding {
	return ResolveWithConfig(candidates, DefaultConfig())
}

// ResolveWithConfig processes candidates with custom configuration
func ResolveWithConfig(candidates []candidate.Candidate, cfg Config) []finding.Finding {
	if len(candidates) == 0 {
		return []finding.Finding{}
	}

	// 1. Parse signals for each candidate
	signals := make([]Signals, len(candidates))
	for i, c := range candidates {
		signals[i] = ParseReason(c.Reason)
	}

	// 2. Find adjacent name pairs (boosts confidence)
	pairedIndices := findPairedIndices(candidates)

	// 3. Score and emit findings
	findings := make([]finding.Finding, 0, len(candidates))

	for i, c := range candidates {
		sig := signals[i]
		isPaired := pairedIndices[i]

		confidence := determineConfidence(sig, isPaired)

		// Skip heuristic confidence if configured to be conservative
		if confidence == finding.ConfidenceHeuristic && !cfg.IncludeHeuristic {
			continue
		}

		findings = append(findings, finding.Finding{
			Type:       finding.EntityName,
			Start:      c.Start,
			End:        c.End,
			Text:       c.Text,
			Detector:   "resolver.name",
			Confidence: confidence,
			Metadata:   buildMetadata(sig, isPaired),
		})
	}

	return findings
}

// findPairedIndices returns a map of candidate indices that are part of adjacent name pairs
// Two candidates are adjacent if they appear next to each other in the token stream
func findPairedIndices(candidates []candidate.Candidate) map[int]bool {
	paired := make(map[int]bool)

	for i := 0; i < len(candidates)-1; i++ {
		curr, next := candidates[i], candidates[i+1]

		// Adjacent if tokens are consecutive (allowing for whitespace token between)
		// EndToken is exclusive, so curr.EndToken should be close to next.StartToken
		if isAdjacent(curr, next) {
			paired[i] = true
			paired[i+1] = true
		}
	}

	return paired
}

// isAdjacent checks if two candidates are adjacent in the token stream
// Allows for 1-2 tokens between (typically whitespace)
func isAdjacent(a, b candidate.Candidate) bool {
	gap := b.StartToken - a.EndToken
	return gap >= 0 && gap <= 2
}

// determineConfidence applies confidence rules based on signals and pairing
func determineConfidence(sig Signals, isPaired bool) finding.Confidence {
	// Deterministic cases - always redact
	if sig.HasTitle {
		return finding.ConfidenceDeterministic
	}

	// Non-ambiguous name, capitalized, mid-sentence
	if !sig.IsAmbiguous && sig.IsCapitalized && !sig.IsSentInit {
		return finding.ConfidenceDeterministic
	}

	// Paired names (e.g., "Will Smith") - both get boosted
	if isPaired && (sig.IsFirstName || sig.IsLastName) {
		return finding.ConfidenceDeterministic
	}

	// Candidate cases - high confidence
	// Mid-sentence capitalized (even if ambiguous)
	if sig.IsCapitalized && !sig.IsSentInit {
		return finding.ConfidenceCandidate
	}

	// Paired but only ambiguous
	if isPaired {
		return finding.ConfidenceCandidate
	}

	// In multiple name lists (e.g., "Thomas" is both first and last name)
	if sig.InMultipleLists() {
		return finding.ConfidenceCandidate
	}

	// Heuristic - low confidence, may need NER
	// Typically: ambiguous + sent_init + no other positive signals
	return finding.ConfidenceHeuristic
}

// buildMetadata creates metadata map for a finding
func buildMetadata(sig Signals, isPaired bool) map[string]string {
	meta := make(map[string]string)

	if sig.IsFirstName {
		meta["firstname"] = "true"
	}
	if sig.IsLastName {
		meta["lastname"] = "true"
	}
	if sig.IsAmbiguous {
		meta["ambiguous"] = "true"
	}
	if sig.IsSentInit {
		meta["sent_init"] = "true"
	}
	if sig.HasTitle {
		meta["has_title"] = "true"
	}
	if isPaired {
		meta["paired"] = "true"
	}

	return meta
}
