package stringslices

import "strings"

func ToLower(s []string) []string {
	res := make([]string, len(s))
	for i, v := range s {
		res[i] = strings.ToLower(v)
	}
	return res
}
