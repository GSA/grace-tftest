package shared

import (
	"strings"
)

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

// Filter is an interface for filtering interface{} objects
type Filter func(interface{}) bool

// GenericFilter executes all filters provided against each item provided
// returning the remaining items
func GenericFilter(filters []Filter, items []interface{}) (result []interface{}) {
	Debugf("len(items) = %d, len(filters) = %d\n", len(items), len(filters))
outer:
	for x, item := range items {
		Debugf("items(%d):\n", x)
		Dump(item)
		for xx, f := range filters {
			if !f(item) {
				continue outer
			}
			Debugf("items(%d) matched filters(%d)\n", x, xx)
		}
		Debugf("storing items(%d)\n", x)
		result = append(result, item)
	}
	Dump(result)
	return
}
