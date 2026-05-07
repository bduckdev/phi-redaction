package detectors_test

import (
	"testing"

	"phi-redactor/internal/detectors"
	"phi-redactor/internal/finding"

	"github.com/google/go-cmp/cmp"
)

func TestFindSSNs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTexts []string
	}{
		{
			name:      "standard_format",
			input:     "SSN: 123-45-6789",
			wantTexts: []string{"123-45-6789"},
		},
		{
			name:      "multiple_ssns",
			input:     "SSN 123-45-6789 and 987-65-4321",
			wantTexts: []string{"123-45-6789", "987-65-4321"},
		},
		{
			name:      "no_ssn",
			input:     "Phone: 555-1234",
			wantTexts: nil,
		},
		{
			name:      "embedded_in_text",
			input:     "Patient SSN is 111-22-3333 on file",
			wantTexts: []string{"111-22-3333"},
		},
		{
			name:      "no_dashes_no_match",
			input:     "123456789",
			wantTexts: nil, // conservative - require dashes
		},
		{
			name:      "partial_match_too_long",
			input:     "123-45-67890",
			wantTexts: nil, // too many digits at end
		},
		{
			name:      "at_start_of_text",
			input:     "123-45-6789 is the SSN",
			wantTexts: []string{"123-45-6789"},
		},
		{
			name:      "at_end_of_text",
			input:     "SSN is 123-45-6789",
			wantTexts: []string{"123-45-6789"},
		},
		{
			name:      "empty_input",
			input:     "",
			wantTexts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := detectors.FindSSNs(tt.input)

			var gotTexts []string
			for _, f := range findings {
				gotTexts = append(gotTexts, f.Text)
				// Verify finding properties
				if f.Type != finding.EntitySSN {
					t.Errorf("expected type SSN, got %s", f.Type)
				}
				if f.Confidence != finding.ConfidenceDeterministic {
					t.Errorf("expected deterministic confidence, got %s", f.Confidence)
				}
				if f.Detector != "regex.ssn" {
					t.Errorf("expected detector regex.ssn, got %s", f.Detector)
				}
			}

			if diff := cmp.Diff(tt.wantTexts, gotTexts); diff != "" {
				t.Errorf("FindSSNs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
