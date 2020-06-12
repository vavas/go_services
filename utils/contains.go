package utils

// ContainsString uses to determine if a string is exists in a slice
func ContainsString(el string, arr []string) bool {
	for _, e := range arr {
		if e == el {
			return true
		}
	}
	return false
}
