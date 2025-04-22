package sliceutils_test

import (
	"github.com/habiliai/agentruntime/internal/sliceutils"
	"testing"
)

func TestRandomSampleN(t *testing.T) {
	t.Run("Given small slice, when sampling n elements bigger than slice, then return all elements", func(t *testing.T) {
		slice := []int{1, 2, 3}
		n := 5
		result := sliceutils.RandomSampleN(slice, n)
		if len(result) != len(slice) {
			t.Errorf("expected %d elements, got %d", len(slice), len(result))
		}
	})

	t.Run("Given big slice, when smapling samller than slice, then return n elements", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		n := 3
		result := sliceutils.RandomSampleN(slice, n)
		if len(result) != n {
			t.Errorf("expected %d elements, got %d", n, len(result))
		}
	})

	t.Run("Given a slice, when sampling n elements, then return n unique elements", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		n := 3
		result := sliceutils.RandomSampleN(slice, n)
		unique := make(map[int]struct{})
		for _, v := range result {
			unique[v] = struct{}{}
		}
		if len(unique) != n {
			t.Errorf("expected %d unique elements, got %d", n, len(unique))
		}
	})
}
