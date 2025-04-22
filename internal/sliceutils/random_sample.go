package sliceutils

import "math/rand/v2"

func RandomSampleN[T any](slice []T, n int) []T {
	res := make([]T, 0, n)
	selected := map[int]struct{}{}
	for i := 0; i < n; i++ {
		for {
			idx := rand.IntN(len(slice))
			if _, ok := selected[idx]; !ok {
				continue
			}
			selected[idx] = struct{}{}
			res = append(res, slice[idx])
			break
		}
	}

	return res
}
