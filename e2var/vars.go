package e2var

// MustStringValue
func MustStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
