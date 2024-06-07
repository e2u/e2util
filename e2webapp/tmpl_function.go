package e2webapp

func until(start, end int) []int {
	var result []int
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}

func add(i, j int) int {
	return i + j
}
