package pir

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

// Test GulliverPIR correctness on DB with short entries.
func TestGulliverPIR(t *testing.T) {
	N := uint64(1 << 10)
	d := uint64(1 << 24)
	logQ := uint64(32)
	logq := uint64(28)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, N, logQ, logq)
	DB := MakeRandomDB(d, uint64(math.Log2(float64(p.P))), &p)
	for i := uint64(0); i < 2; i++ {
		index := RandInt(big.NewInt(int64(d))).Uint64()
		fmt.Printf("Retrieving index %d \n", index)
		RunPIR(&pir, DB, p, index)
	}

}
