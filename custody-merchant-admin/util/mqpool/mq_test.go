package mqpool

import (
	"fmt"
	"testing"
)

func TestRuns(t *testing.T) {
	fmt.Println("开始")
	rund()
}

func TestRunc(t *testing.T) {
	initConsumerabbitmq()
	Consume()
}
