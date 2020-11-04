package f_test

import (
	"testing"
	"time"

	"github.com/angenalZZZ/gofunc/f"
)

func TestNewTimeFrames(t *testing.T) {
	frames := f.NewTimeFrames(time.Now(), time.Now().Add(time.Hour).Add(20*time.Second), 15*time.Second)
	for i, frame := range frames {
		t.Log(i+1, frame)
	}
}

func TestNewTimeFrame(t *testing.T) {
	var configMod time.Time
	t.Log(configMod.Year() == 1)
}
