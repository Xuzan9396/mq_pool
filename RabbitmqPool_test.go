package mq_pool

import (
	"math"
	"testing"
)

func Test_rPush(t *testing.T) {
	t.Log(testPush(0))
	t.Log(RoundRobin(0, 2))
	t.Log(RoundRobin(1, 2))
	t.Log(RoundRobin(2, 2))
	t.Log(RoundRobin(3, 2))
	t.Log(math.Pow(2, float64(5)))
}

func testPush(sendTime int) string {
	if sendTime >= 5 {
		return "超过5次了"
	}

	sendTime++
	return testPush(sendTime)
}

func RoundRobin(cIndex, max int32) int32 {
	if max == 0 {
		return 0
	}
	return (cIndex) % max
}
