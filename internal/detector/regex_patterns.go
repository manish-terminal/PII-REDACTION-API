package detector

import "regexp"

type RegexPattern struct {
	Name       string
	Pattern    *regexp.Regexp
	Validator  func(match string) bool
	Confidence float64
}

// Locale-specific pattern maps.
var localePatterns = map[string][]RegexPattern{
	"en-US": {
		{
			Name:       "EMAIL",
			Pattern:    regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
			Confidence: 0.99,
		},
		{
			Name:       "SSN",
			Pattern:    regexp.MustCompile(`\b(\d{3}-\d{2}-\d{4})\b`),
			Validator:  validateSSN,
			Confidence: 0.95,
		},
		{
			Name:       "CREDIT_CARD",
			Pattern:    regexp.MustCompile(`\b(?:\d[ -]?){13,19}\b`),
			Validator:  luhnCheck,
			Confidence: 0.97,
		},
		{
			Name:       "PHONE_US",
			Pattern:    regexp.MustCompile(`\b(?:\+?1[\s\-.]?)?\(?\d{3}\)?[\s\-.]?\d{3}[\s\-.]?\d{4}\b`),
			Confidence: 0.90,
		},
		{
			Name:       "IP_ADDRESS",
			Pattern:    regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			Validator:  validateIPv4,
			Confidence: 0.92,
		},
	},
}

func luhnCheck(cardNumber string) bool {
	digits := []int{}
	for _, r := range cardNumber {
		if r >= '0' && r <= '9' {
			digits = append(digits, int(r-'0'))
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	sum := 0
	alternate := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if alternate {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alternate = !alternate
	}
	return sum%10 == 0
}

func validateSSN(match string) bool {
	// Simple validation: should not start with 000, 666, or 900-999.
	// This is a basic example.
	return !regexp.MustCompile(`^(000|666|9\d{2})`).MatchString(match)
}

func validateIPv4(match string) bool {
	// Basic regex already catches 4 groups of digits.
	// We can add logic to check if each octet is <= 255.
	return true // Placeholder
}
