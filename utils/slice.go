package utils

// IndexOf returns the index of str in the slice; -1 if not found
func IndexOf(strs []string, str string) int {
	for i := range strs {
		if strs[i] == str {
			return i
		}
	}

	return -1
}
