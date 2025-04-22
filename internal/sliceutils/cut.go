package sliceutils

func Cut[T any](slice []T, start, end int) []T {
	if len(slice) == 0 {
		return slice
	}

	if start < 0 {
		start = len(slice) - 1 + start
	}
	if end < 0 {
		end = len(slice) - 1 + end
	}

	return slice[max(start, 0):min(end, len(slice))]
}
