package runpipeline

import (
	"strings"

	"fluffy-giggle/core/internal/domain/port"
	"fluffy-giggle/core/internal/domain/vo"
)

const minFormLength = 3

// Egress is the last check before anything reaches the cloud: every spelling the
// vault knows, every number without separators, and the deterministic detectors
// again. Dates are left to the vault: the input's dates are already cut, a date in
// a draft is the model's own.
func Egress(payload string, vault *vo.Vault, detector port.Detector) []vo.Finding {
	var findings []vo.Finding
	seen := map[string]bool{}
	add := func(t vo.PIIType, value string) {
		key := string(t) + ":" + value
		if seen[key] {
			return
		}
		seen[key] = true
		findings = append(findings, vo.Finding{Type: t, Value: value})
	}

	lower := strings.ToLower(payload)
	digits := vo.Digits(payload)
	for _, entry := range vault.Entries() {
		for _, form := range entry.Forms {
			if len([]rune(form)) >= minFormLength && strings.Contains(lower, strings.ToLower(form)) {
				add(entry.Type, form)
			}
		}
		if entry.Type.Numeric() {
			if d := vo.Digits(entry.Canonical); len(d) >= 6 && strings.Contains(digits, d) {
				add(entry.Type, entry.Canonical)
			}
		}
	}
	for _, span := range detector.Detect(payload) {
		if span.Type == vo.PIIDate {
			continue
		}
		add(span.Type, span.Text)
	}

	return findings
}

// countCanaries tells how many planted values reached the cloud, judged on what was
// actually sent rather than on what the vault believes.
func countCanaries(canaries, payloads []string) int {
	joined := strings.ToLower(strings.Join(payloads, "\n"))
	joinedDigits := vo.Digits(joined)
	leaked := 0
	for _, canary := range canaries {
		if canary == "" {
			continue
		}
		if strings.Contains(joined, strings.ToLower(canary)) {
			leaked++
			continue
		}
		if d := vo.Digits(canary); len(d) >= 6 && strings.Contains(joinedDigits, d) {
			leaked++
		}
	}
	return leaked
}
