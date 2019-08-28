package shared

import "strings"

// StringSliceEqual ... validates that two slice of strings are equal
// where 'a' is the authoritative slice
func StringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		Debugf("match failed at length check -> %d != %d\n", len(a), len(b))
		return false
	}
outer:
	for i := 0; i < len(a); i++ {
		for r := 0; r < len(b); r++ {
			if strings.EqualFold(a[i], b[r]) {
				continue outer
			}
		}
		Debugf("matched failed at [a(%d): %s]\n", i, a[i])
		return false
	}
	return true
}
