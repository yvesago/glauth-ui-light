package helpers

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFailLimiter(t *testing.T) {
	var s Status
	// fmt.Printf(" json string: %+v\n", s.ToJSONStr())

	/**
	Test unlock with counter
	**/
	for i := 1; i < 3; i++ {
		s = FailLimiter(s, 10)
		// fmt.Printf("%+v\n", s)
		assert.Equal(t, true, (s.Lock == false && s.Count == i), "2 fails doesn't lock")
		time.Sleep(1 * time.Second)
	}

	/**
	Test method ToJSONStr, fonction StrToStatus
	**/
	str := s.ToJSONStr()
	fmt.Printf(" json string: %+v\n", str)
	o := StrToStatus(str)
	fmt.Printf(" json to Status: %+v\n", StrToStatus(str))
	assert.Equal(t, false, o.Lock, "convert string to object")

	o = StrToStatus("")
	fmt.Printf(" json to Status: %+v\n", StrToStatus(""))
	assert.Equal(t, false, o.Lock, "default object Lock value")
	assert.Equal(t, int64(0), o.LastSeen, "default object LastSeen value")
	assert.Equal(t, 0, o.Count, "default object Count value")

	/**
	Test unlock && reset count, 11 sec later
	**/
	fmt.Println("wait 11s ...")
	time.Sleep(11 * time.Second)
	for i := 1; i < 6; i++ {
		s = FailLimiter(s, 10)
		fmt.Printf("%+v\n", s)
		if i <= 3 {
			assert.Equal(t, true, (s.Lock == false && s.Count == i), "Count reset && 3 fails doesn't lock")
		} else {
			assert.Equal(t, true, (s.Lock == true && s.Count == 0), "Lock && count = 0")
		}
		time.Sleep(1 * time.Second)
	}
}
