package e2var

// MustStringValue
func MustStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func Val2Pointer[T any](i T) *T {
	return &i
}
