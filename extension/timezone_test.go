package extension

import (
	"fmt"
	"testing"
)

func TestTimezone(t *testing.T) {
	city := "shenzhen"
	if GetTimezoneByAddress(city) != "Asia/Shanghai" {
		t.Fail()
	}

	fmt.Println(UTCNowOnAddress(city))
}
