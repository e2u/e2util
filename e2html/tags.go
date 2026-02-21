// https://www.w3schools.com/tags/tag_comment.asp

package e2html

import (
	"fmt"
	"regexp"
)

// validDoctypePattern matches safe doctype identifiers
// Only allows alphanumeric, spaces, and dashes
var validDoctypePattern = regexp.MustCompile(`^[a-zA-Z0-9\-\s]+$`)

// Doctype generates a DOCTYPE declaration.
// The input is validated to prevent XSS attacks.
// Returns a safe default if the input contains invalid characters.
func Doctype(t string) string {
	// Validate input - only allow safe characters
	if !validDoctypePattern.MatchString(t) {
		// Return safe default for invalid input
		t = "html"
	}
	return fmt.Sprintf("<!DOCTYPE %s>", t)
}

func Div(args ...any) TAG {
	return T("div", args...)
}

// ....
