package detectors

import (
	"regexp"

	"phi-redactor/internal/finding"
)

// SSN pattern: 123-45-6789 (standard format with dashes)
var ssnRE = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)

// FindSSNs detects Social Security Numbers in text
func FindSSNs(text string) []finding.Finding {
	matches := ssnRE.FindAllStringIndex(text, -1)

	findings := make([]finding.Finding, 0, len(matches))
	for _, m := range matches {
		findings = append(findings, finding.Finding{
			Type:       finding.EntitySSN,
			Start:      m[0],
			End:        m[1],
			Text:       text[m[0]:m[1]],
			Detector:   "regex.ssn",
			Confidence: finding.ConfidenceDeterministic,
		})
	}

	return findings
}
