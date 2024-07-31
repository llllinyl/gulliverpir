package pir

import (
	"fmt"
	"math"
	"time"
)

// Helper function to print the elapsed time since start.
func printTime(start time.Time) time.Duration {
	elapsed := time.Since(start)
	fmt.Printf("\tElapsed: %s\n", elapsed)
	return elapsed
}

// Helper function to print the transfer rate in MB/s.
func printRate(p Params, elapsed time.Duration, batch_sz int) float64 {
	rate := math.Log2(float64((p.P))) * float64(p.L*p.M) * float64(batch_sz) /
		float64(8*1024*1024*elapsed.Seconds())
	fmt.Printf("\tRate: %f MB/s\n", rate)
	return rate
}

// Helper function to calculate communication size in KB.
func calculateCommunicationSize(size uint64, logMod uint64) float64 {
	return float64(size) * float64(logMod) / (8.0 * 1024.0)
}
