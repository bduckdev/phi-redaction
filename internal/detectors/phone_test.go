package detectors_test

import (
	"testing"

	"phi-redactor/internal/detectors"
	"phi-redactor/internal/finding"

	"github.com/google/go-cmp/cmp"
)

func TestFindPhones(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTexts []string
	}{
		{
			name:      "parens_format",
			input:     "Call (555) 123-4567",
			wantTexts: []string{"(555) 123-4567"},
		},
		{
			name:      "parens_no_space",
			input:     "Call (555)123-4567",
			wantTexts: []string{"(555)123-4567"},
		},
		{
			name:      "dashed_format",
			input:     "Phone: 555-123-4567",
			wantTexts: []string{"555-123-4567"},
		},
		{
			name:      "dotted_format",
			input:     "Contact: 555.123.4567",
			wantTexts: []string{"555.123.4567"},
		},
		{
			name:      "spaced_format",
			input:     "Number: 555 123 4567",
			wantTexts: []string{"555 123 4567"},
		},
		{
			name:      "with_country_code",
			input:     "Call +1-555-123-4567",
			wantTexts: []string{"+1-555-123-4567"},
		},
		{
			name:      "country_code_dots",
			input:     "Call +1.555.123.4567",
			wantTexts: []string{"+1.555.123.4567"},
		},
		{
			name:      "multiple_phones",
			input:     "Home: 555-111-2222, Work: 555-333-4444",
			wantTexts: []string{"555-111-2222", "555-333-4444"},
		},
		{
			name:      "no_phone",
			input:     "Order #12345678901",
			wantTexts: nil,
		},
		{
			name:      "ssn_not_phone",
			input:     "SSN: 123-45-6789",
			wantTexts: nil, // SSN format (3-2-4) != phone (3-3-4)
		},
		{
			name:      "embedded_in_text",
			input:     "Reach me at 555-123-4567 anytime",
			wantTexts: []string{"555-123-4567"},
		},
		{
			name:      "at_start",
			input:     "555-123-4567 is my number",
			wantTexts: []string{"555-123-4567"},
		},
		{
			name:      "at_end",
			input:     "My number is 555-123-4567",
			wantTexts: []string{"555-123-4567"},
		},
		{
			name:      "empty_input",
			input:     "",
			wantTexts: nil,
		},
		{
			name:      "mixed_formats",
			input:     "Call (555) 111-2222 or 555.333.4444",
			wantTexts: []string{"(555) 111-2222", "555.333.4444"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := detectors.FindPhones(tt.input)

			var gotTexts []string
			for _, f := range findings {
				gotTexts = append(gotTexts, f.Text)
				// Verify finding properties
				if f.Type != finding.EntityPhone {
					t.Errorf("expected type PHONE, got %s", f.Type)
				}
				if f.Confidence != finding.ConfidenceDeterministic {
					t.Errorf("expected deterministic confidence, got %s", f.Confidence)
				}
				if f.Detector != "regex.phone" {
					t.Errorf("expected detector regex.phone, got %s", f.Detector)
				}
			}

			if diff := cmp.Diff(tt.wantTexts, gotTexts); diff != "" {
				t.Errorf("FindPhones() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
