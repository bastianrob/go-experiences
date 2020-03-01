package filter_test

import (
	"testing"

	"github.com/bastianrob/go-experiences/filter"
)

func BenchmarkDeferredFilterFast(b *testing.B) {
	source := [100]int{}
	for i := 0; i < len(source); i++ {
		source[i] = i + 1
	}
	isMultipliedBy3 := func(num int) bool {
		return num%3 == 0
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		q, _ := filter.DeferredFilter(source, isMultipliedBy3)
		for _ = range q {
		}
	}
}
