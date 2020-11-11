package store

import (
	"fmt"
	"math"
	"testing"
)

func BenchmarkNdbSet(b *testing.B) {
	store := NewNdb(nil)
	defer func() { _ = store.Close() }()

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				_ = store.Set(key, value, &Options{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkNdbGet(b *testing.B) {
	store := NewNdb(nil)
	defer func() { _ = store.Close() }()

	key := "test"
	value := []byte("value")

	_ = store.Set(key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(key)
			}
		})
	}
}
