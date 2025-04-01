package stringslices

import "strings"

func IntersectIgnoreCase(a, b []string) []string {
	m := make(map[string]struct{}, len(a))
	for _, s := range a {
		m[strings.ToLower(s)] = struct{}{}
	}

	var res []string
	for _, s := range b {
		if _, ok := m[strings.ToLower(s)]; ok {
			res = append(res, s)
		}
	}

	return res
}
