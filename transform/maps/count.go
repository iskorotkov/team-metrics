package maps

func Count[K ~string, V any](m map[K][]V) map[K]int {
	counts := make(map[K]int, len(m))
	for k, v := range m {
		counts[k] += len(v)
	}
	return counts
}
