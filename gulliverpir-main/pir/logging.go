package pir

import (
	"fmt"
	"math"
	"time"
)

func printTime(start time.Time) time.Duration {
	elapsed := time.Since(start)
	fmt.Printf("\tElapsed: %s\n", elapsed)
	return elapsed
}

func printRate(p Params, elapsed time.Duration, batch_sz int) float64 {
	rate := math.Log2(float64((p.P))) * float64(p.L*p.M) * float64(batch_sz) /
		float64(8*1024*1024*elapsed.Seconds())
	fmt.Printf("\tRate: %f MB/s\n", rate)
	return rate
}
