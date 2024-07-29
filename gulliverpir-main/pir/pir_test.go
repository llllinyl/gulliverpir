package pir

import (
	"fmt"
	"math/big"
	"testing"
)

// Test GulliverPIR correctness on DB with short entries.
func TestGulliverPIR(t *testing.T) {
	N := uint64(1 << 10)
	d := uint64(1 << 26)
	logQ := uint64(32)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, N, logQ)

	DB := MakeRandomDB(d, uint64(8), &p)

	for i := uint64(0); i < 2; i++ {
		index := RandInt(big.NewInt(int64(d))).Uint64()
		fmt.Printf("Retrieving index %d \n", index)
		RunPIR(&pir, DB, p, index)
	}

}
