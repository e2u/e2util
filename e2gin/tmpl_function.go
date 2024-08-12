package e2gin

func until(start, end int) []int {
	var result []int
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}
