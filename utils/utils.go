// utils package includes some utility functions.
package utils

// SearchStrings returns the smallest index of x in the slice s
// or -1 if there is no x inside slice s.
func SearchStrings(s []string, x string) int {
	for i, ss := range s {
		if ss == x {
			return i
		}
	}

	return -1
}
