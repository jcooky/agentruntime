package sliceutils

import "math/rand/v2"

func RandomSampleN[T any](slice []T, n int) []T {
	n = min(n, len(slice))
	res := make([]T, 0, n)
	indices := rand.Perm(len(slice))

	for i := 0; i < n; i++ {
		res = append(res, slice[indices[i]])
	}

	return res
}
