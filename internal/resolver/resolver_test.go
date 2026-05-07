package resolver_test

import (
	"testing"

	"phi-redactor/internal/candidate"
	"phi-redactor/internal/detectors"
	"phi-redactor/internal/finding"
	"phi-redactor/internal/lexer"
	"phi-redactor/internal/resolver"
	"phi-redactor/internal/token"

	"github.com/google/go-cmp/cmp"
)

func tokenize(input string) []token.Token {
	l := lexer.New(input)
	tokens := make([]token.Token, 0)
	for {
		tok := l.NextToken()
		if tok.Text == "" {
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

// Helper to run full pipeline: tokenize -> detect -> resolve
func resolveText(input string) []finding.Finding {
	tokens := tokenize(input)
	candidates := detectors.FindNameCandidates(tokens)
	return resolver.Resolve(candidates)
}

func TestParseReason(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		want   resolver.Signals
	}{
		{
			name:   "empty",
			reason: "",
			want:   resolver.Signals{},
		},
		{
			name:   "firstname only",
			reason: "firstname",
			want:   resolver.Signals{IsFirstName: true},
		},
		{
			name:   "full signal set",
			reason: "firstname:lastname:capitalized:sent_init:has_title",
			want: resolver.Signals{
				IsFirstName:   true,
				IsLastName:    true,
				IsCapitalized: true,
				IsSentInit:    true,
				HasTitle:      true,
			},
		},
		{
			name:   "ambiguous with context",
			reason: "ambiguous:capitalized:sent_init",
			want: resolver.Signals{
				IsAmbiguous:   true,
				IsCapitalized: true,
				IsSentInit:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver.ParseReason(tt.reason)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseReason() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveConfidence(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTexts []string
		wantConfs []finding.Confidence
	}{
		// === Deterministic cases ===
		{
			name:      "title_prefix_deterministic",
			input:     "Dr. Smith called.",
			wantTexts: []string{"Smith"},
			wantConfs: []finding.Confidence{finding.ConfidenceDeterministic},
		},
		{
			name:      "mid_sentence_firstname_deterministic",
			input:     "I saw James today.",
			wantTexts: []string{"James"},
			wantConfs: []finding.Confidence{finding.ConfidenceDeterministic},
		},
		{
			name:      "mid_sentence_lastname_deterministic",
			input:     "I met Smith yesterday.",
			wantTexts: []string{"Smith"},
			wantConfs: []finding.Confidence{finding.ConfidenceDeterministic},
		},
		{
			name:      "paired_names_mixed_confidence",
			input:     "Will Smith arrived.",
			wantTexts: []string{"Will", "Smith"},
			wantConfs: []finding.Confidence{finding.ConfidenceCandidate, finding.ConfidenceDeterministic}, // Will=ambiguous+paired, Smith=lastname+paired
		},
		{
			name:      "paired_first_last_deterministic",
			input:     "I called John Smith.",
			wantTexts: []string{"John", "Smith"},
			wantConfs: []finding.Confidence{finding.ConfidenceDeterministic, finding.ConfidenceDeterministic},
		},

		// === Candidate cases (high confidence) ===
		{
			name:      "ambiguous_mid_sentence_candidate",
			input:     "I saw Hope today.",
			wantTexts: []string{"Hope"},
			wantConfs: []finding.Confidence{finding.ConfidenceCandidate},
		},
		{
			name:      "both_first_and_last_candidate",
			input:     "I met Thomas yesterday.",
			wantTexts: []string{"Thomas"},
			wantConfs: []finding.Confidence{finding.ConfidenceDeterministic}, // firstname:lastname + mid-sentence
		},

		// === Heuristic cases (low confidence) ===
		{
			name:      "ambiguous_sent_init_heuristic",
			input:     "Hope arrived early.",
			wantTexts: []string{"Hope"},
			wantConfs: []finding.Confidence{finding.ConfidenceHeuristic},
		},
		{
			name:      "faith_will_prevail_heuristic",
			input:     "Faith will prevail.",
			wantTexts: []string{"Faith"},
			wantConfs: []finding.Confidence{finding.ConfidenceHeuristic},
		},
		{
			name:      "grace_subject_heuristic",
			input:     "Grace will help.",
			wantTexts: []string{"Grace"},
			wantConfs: []finding.Confidence{finding.ConfidenceHeuristic},
		},

		// === Mixed cases ===
		{
			name:      "rose_smith_paired",
			input:     "Rose Smith called.",
			wantTexts: []string{"Rose", "Smith"},
			wantConfs: []finding.Confidence{finding.ConfidenceCandidate, finding.ConfidenceDeterministic},
		},
		{
			name:      "multiple_names_sentence",
			input:     "James and Mary talked.",
			wantTexts: []string{"James", "Mary"},
			wantConfs: []finding.Confidence{finding.ConfidenceHeuristic, finding.ConfidenceDeterministic}, // James=sent_init, Mary=mid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := resolveText(tt.input)

			var gotTexts []string
			var gotConfs []finding.Confidence
			for _, f := range findings {
				gotTexts = append(gotTexts, f.Text)
				gotConfs = append(gotConfs, f.Confidence)
			}

			if diff := cmp.Diff(tt.wantTexts, gotTexts); diff != "" {
				t.Errorf("texts mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantConfs, gotConfs); diff != "" {
				t.Errorf("confidences mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveWithConfigExcludeHeuristic(t *testing.T) {
	tokens := tokenize("Hope arrived early.")
	candidates := detectors.FindNameCandidates(tokens)

	// With heuristics included (default)
	withHeuristic := resolver.Resolve(candidates)
	if len(withHeuristic) != 1 {
		t.Errorf("expected 1 finding with heuristics, got %d", len(withHeuristic))
	}

	// With heuristics excluded
	cfg := resolver.Config{IncludeHeuristic: false}
	withoutHeuristic := resolver.ResolveWithConfig(candidates, cfg)
	if len(withoutHeuristic) != 0 {
		t.Errorf("expected 0 findings without heuristics, got %d", len(withoutHeuristic))
	}
}

func TestResolvePairedNames(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPaired []bool
	}{
		{
			name:       "adjacent_first_last",
			input:      "John Smith arrived.",
			wantPaired: []bool{true, true},
		},
		{
			name:       "non_adjacent_names",
			input:      "John and Smith talked.",
			wantPaired: []bool{false, false},
		},
		{
			name:       "two_adjacent_names_with_gap",
			input:      "Mary Elizabeth Smith called.",
			wantPaired: []bool{true, true, true}, // Mary, Elizabeth, Smith all paired
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := resolveText(tt.input)

			var gotPaired []bool
			for _, f := range findings {
				gotPaired = append(gotPaired, f.Metadata["paired"] == "true")
			}

			if diff := cmp.Diff(tt.wantPaired, gotPaired); diff != "" {
				t.Errorf("paired mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveMetadata(t *testing.T) {
	findings := resolveText("Dr. Smith called.")

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.Metadata["has_title"] != "true" {
		t.Error("expected has_title metadata")
	}
	if f.Metadata["lastname"] != "true" {
		t.Error("expected lastname metadata")
	}
	if f.Detector != "resolver.name" {
		t.Errorf("expected detector=resolver.name, got %s", f.Detector)
	}
	if f.Type != finding.EntityName {
		t.Errorf("expected type=NAME, got %s", f.Type)
	}
}

func TestResolveEmptyInput(t *testing.T) {
	findings := resolver.Resolve([]candidate.Candidate{})
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty input, got %d", len(findings))
	}
}
