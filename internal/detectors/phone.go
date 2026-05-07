package detectors

import (
	"regexp"

	"phi-redactor/internal/finding"
)

// Phone patterns for US formats
var phonePatterns = []*regexp.Regexp{
	// (555) 123-4567 or (555) 123 4567 or (555)123-4567
	regexp.MustCompile(`\(\d{3}\)\s?\d{3}[-.\s]?\d{4}`),
	// 555-123-4567 or 555.123.4567 or 555 123 4567
	regexp.MustCompile(`\b\d{3}[-.\s]\d{3}[-.\s]\d{4}\b`),
	// +1-555-123-4567 or +1 555 123 4567 or +1.555.123.4567
	regexp.MustCompile(`\+1[-.\s]?\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`),
}

// FindPhones detects phone numbers in text
func FindPhones(text string) []finding.Finding {
	var findings []finding.Finding

	for _, re := range phonePatterns {
		matches := re.FindAllStringIndex(text, -1)
		for _, m := range matches {
			findings = append(findings, finding.Finding{
				Type:       finding.EntityPhone,
				Start:      m[0],
				End:        m[1],
				Text:       text[m[0]:m[1]],
				Detector:   "regex.phone",
				Confidence: finding.ConfidenceDeterministic,
			})
		}
	}

	// Remove overlapping matches, keeping the longest one
	findings = removeOverlappingPhones(findings)

	return findings
}

// removeOverlappingPhones removes overlapping phone findings, keeping the longest match
func removeOverlappingPhones(findings []finding.Finding) []finding.Finding {
	if len(findings) <= 1 {
		return findings
	}

	var result []finding.Finding
	for _, f := range findings {
		overlaps := false
		for i, existing := range result {
			// Check if f overlaps with existing
			if f.Start < existing.End && f.End > existing.Start {
				overlaps = true
				// Keep the longer match
				if (f.End - f.Start) > (existing.End - existing.Start) {
					result[i] = f
				}
				break
			}
		}
		if !overlaps {
			result = append(result, f)
		}
	}

	return result
}
