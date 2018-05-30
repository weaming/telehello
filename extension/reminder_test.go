package extension

import (
	"fmt"
	"testing"
)

func TestWhen(t *testing.T) {
	for _, text := range []string{
		"remind me do someting at 2:00 pm friday",
		"remind me to wash on next wednesday at 2:25 p.m.",
	} {
		result, err := ParseHuman(text)
		if err != nil {
			t.Error(err)
		} else {
			fmt.Printf("I will remind you at %v: %v\n", result.Time, result.Source)
		}
	}
}
