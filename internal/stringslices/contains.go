package stringslices

import "strings"

func ContainsIgnoreCase(a []string, s string) bool {
	for _, v := range a {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
