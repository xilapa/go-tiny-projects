package lfucache

import (
	"fmt"
	"testing"
)

func BenchmarkLinkedListItemCast(b *testing.B) {
	c := New(2)
	c.Add("key one", 1)
	c.Add("key two", 2)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.Add(fmt.Sprintf("key %d", n), 3)
	}
}
