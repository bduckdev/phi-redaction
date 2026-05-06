// Email detector is fine as just being a regex
package detectors

import (
	"regexp"

	"phi-redactor/internal/finding"
)

var emailRE = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)

func FindEmails(text string) []finding.Finding {
	matches := emailRE.FindAllStringIndex(text, -1)

	findings := make([]finding.Finding, 0, len(matches))
	for _, m := range matches {
		findings = append(findings, finding.Finding{
			Type:     finding.EntityEmail,
			Start:    m[0],
			End:      m[1],
			Detector: "regex.email",
		})
	}

	return findings
}
