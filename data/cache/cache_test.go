package cache

import "testing"

func TestNewChain(t *testing.T) {
	t.Log(new(Cache).getMD5Key(123))
}
