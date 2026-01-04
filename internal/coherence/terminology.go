package coherence

import "strings"

// NormalizeTerminology converts variant terms to their canonical form.
func NormalizeTerminology(text string, termRegistry map[string]string) string {
	for variant, canonical := range termRegistry {
		text = strings.ReplaceAll(strings.ToLower(text), variant, canonical)
	}
	return text
}

// CheckTerminologyConsistency ensures terminology is used consistently across outputs.
func CheckTerminologyConsistency(outputs []OutputContext) (consistent bool, issues []string) {
	consistent = true
	issues = []string{}

	termFreq := make(map[string]int)

	for _, out := range outputs {
		for _, term := range out.Terms {
			termFreq[term]++
		}
	}

	// Check if variant terms are used inconsistently
	if termFreq["arm"] > 0 && termFreq["activate"] > 0 {
		consistent = false
		issues = append(issues, "term_inconsistency: both 'arm' and 'activate' used")
	}
	if termFreq["disarm"] > 0 && termFreq["turn off"] > 0 {
		consistent = false
		issues = append(issues, "term_inconsistency: both 'disarm' and 'turn off' used")
	}

	return consistent, issues
}
