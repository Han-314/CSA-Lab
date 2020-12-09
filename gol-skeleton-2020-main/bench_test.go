package main

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

// Naming convention
// func BenchmarkSequential(b *testing.B) {
// 	for size := start; size <= end; size *= 2 {
// 		b.Run(fmt.Sprint(size), func(b *testing.B) {
// 			os.Stdout = nil // Disable all program output apart from benchmark results
// 			for i := 0; i < b.N; i++ {
// 				unsorted := random(size)
// 				b.StartTimer()
// 				mergeSort(unsorted)
// 				b.StopTimer()
// 			}
// 		})
// 	}
// }
func BenchmarkGol(b *testing.B) {
	os.Stdout = nil
	tests := []gol.Params{
		//{ImageWidth: 16, ImageHeight: 16},
		{ImageWidth: 64, ImageHeight: 64},
		//{ImageWidth: 512, ImageHeight: 512},
	}
	for _, p := range tests {
		for _, turns := range []int{1000} {
			p.Turns = turns

			for threads := 1; threads <= 8; threads++ {
				p.Threads = threads
				testName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
				b.Run(testName, func(b *testing.B) {
					os.Stdout = nil
					events := make(chan gol.Event)
					for i := 0; i < b.N; i++ {
						// events := make(chan gol.Event)
						// keyPresses := make(chan rune, 10)
						b.StartTimer()
						gol.Run(p, events, nil)
						b.StopTimer()
					}

				})
			}
		}
	}
}

// func BenchmarkStruckChan(b *testing.B) {
// 	ch := make(chan struct{})
// 	go func() {
// 		for {
// 			<-ch
// 		}
// 	}()
// 	for i := 0; i < b.N; i++ {
// 		ch <- struct{}{}
// 	}
// }
