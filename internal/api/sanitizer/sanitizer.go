package sanitizer

import (
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	policy      *bluemonday.Policy
	policyOnce  sync.Once
)

// getPolicy returns the strict sanitization policy.
// We use a strict policy that removes all HTML tags to prevent XSS.
// If you need to allow rich text, change StrictPolicy() to UGCPolicy().
func getPolicy() *bluemonday.Policy {
	policyOnce.Do(func() {
		policy = bluemonday.StrictPolicy()
	})
	return policy
}

// SanitizeString removes all HTML tags and trims whitespace.
func SanitizeString(s string) string {
	sanitized := getPolicy().Sanitize(s)
	return strings.TrimSpace(sanitized)
}

// Sanitizable interface should be implemented by DTOs that need sanitization.
type Sanitizable interface {
	Sanitize()
}
