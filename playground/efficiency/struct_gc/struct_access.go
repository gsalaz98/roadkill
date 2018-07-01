package main

// We want to identify in this test the performance differences by accessing a struct by array value vs initializing it at
// the end and inserting it at that time. We use the orderbook.Delta structure to benchmark these both methods.

import (
	"fmt"
	"time"

	"github.com/gsalaz98/roadkill/orderbook"
)

func average(looptime []uint64) float64 {
	var total uint64 = 0
	for _, v := range looptime {
		total += v
	}
	return float64(total) / float64(len(looptime))
}

func minMax(array []uint64) (uint64, uint64) {
	var max = array[0]
	var min = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func main() {
	loop1 := make([]uint64, 100)
	loop2 := make([]uint64, 100)

	for x := 0; x < 100; x++ {
		start := time.Now()
		for i := 0; i < 1; i++ {
			deltas := make([]orderbook.Delta, 10)
			for j := 0; j < 10; j++ {
				deltas[j] = orderbook.Delta{
					TimeDelta: uint64(time.Now().UnixNano()),
					Seq:       0,
					Event:     0,
					Price:     0,
					Size:      0,
				}
				deltas[j].Seq = 100
				deltas[j].Event = 100
				deltas[j].Price = 50.43
				deltas[j].Size = 0.04
			}
		}

		end := time.Now().Sub(start)
		loop1[x] = uint64(end / time.Nanosecond)
		fmt.Println("array access: ", end)

		start = time.Now()

		for i := 0; i < 1; i++ {
			deltas := make([]orderbook.Delta, 10)
			for j := 0; j < 10; j++ {
				time := uint64(time.Now().UnixNano())
				var seq uint64 = 100
				var event uint8 = 100
				price := 50.43
				size := 0.04

				deltas[j] = orderbook.Delta{
					TimeDelta: time,
					Seq:       seq,
					Event:     event,
					Price:     price,
					Size:      size,
				}
			}
		}

		end = time.Now().Sub(start)
		loop2[x] = uint64(end / time.Nanosecond)
		fmt.Println("struct jit init: ", end)
		fmt.Println("==================")
	}
	loop1Min, loop1Max := minMax(loop1)
	fmt.Printf("Loop1 Average: %f, min: %d, max: %d, std: None\n", average(loop1), loop1Min, loop1Max)

	loop2Min, loop2Max := minMax(loop2)
	fmt.Printf("Loop2 Average: %f, min: %d, max: %d, std: None\n", average(loop2), loop2Min, loop2Max)
}

// gccgo produces faster times for jit struct init, but is severely limited by array accessing.
// go compiler has no difference between these two methods. Perhaps optimizing for gccgo might be a mistake?
