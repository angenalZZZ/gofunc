package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
	"time"
)

func TestNewTimeFrames(t *testing.T) {
	frames := f.NewTimeFrames(time.Now(), time.Now().Add(time.Hour).Add(20*time.Second), 15*time.Second)
	for i, frame := range frames {
		t.Log(i+1, frame)
	}
}
