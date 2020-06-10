package f

import (
	"testing"
	"time"
)

func TestNewTimeFrames(t *testing.T) {
	frames := NewTimeFrames(time.Now(), time.Now().Add(time.Hour).Add(20*time.Second), 15*time.Second)
	for i, frame := range frames {
		t.Log(i+1, frame)
	}
}
