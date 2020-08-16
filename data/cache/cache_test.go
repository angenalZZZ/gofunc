package cache

import "testing"

func TestNewChain(t *testing.T) {
	t.Log(new(Cache).GetType())
}
