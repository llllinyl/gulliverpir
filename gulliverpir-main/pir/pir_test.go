package pir

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const LOGQ = uint64(32)
const SEC_PARAM = uint64(1 << 10)

// Test that DB packing methods are correct, when each database entry is ~ 1 Z_p elem.
func TestDBMediumEntries(t *testing.T) {
	N := uint64(4)
	d := uint64(9)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

	vals := []uint64{1, 2, 3, 4}
	DB := MakeDB(N, d, &p, vals)
	if DB.Info.Packing != 1 || DB.Info.Ne != 1 {
		panic("Should not happen.")
	}

	for i := uint64(0); i < N; i++ {
		if DB.GetElem(i) != (i + 1) {
			panic("Failure")
		}
	}
}

// Test that DB packing methods are correct, when multiple database entries fit in 1 Z_p elem.
func TestDBSmallEntries(t *testing.T) {
	N := uint64(4)
	d := uint64(3)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

	vals := []uint64{1, 2, 3, 4}
	DB := MakeDB(N, d, &p, vals)
	if DB.Info.Packing <= 1 || DB.Info.Ne != 1 {
		panic("Should not happen.")
	}

	for i := uint64(0); i < N; i++ {
		if DB.GetElem(i) != (i + 1) {
			panic("Failure")
		}
	}
}

// Test that DB packing methods are correct, when each database entry requires multiple Z_p elems.
func TestDBLargeEntries(t *testing.T) {
	N := uint64(4)
	d := uint64(12)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

	vals := []uint64{1, 2, 3, 4}
	DB := MakeDB(N, d, &p, vals)
	if DB.Info.Packing != 0 || DB.Info.Ne <= 1 {
		panic("Should not happen.")
	}

	for i := uint64(0); i < N; i++ {
		if DB.GetElem(i) != (i + 1) {
			panic("Failure")
		}
	}
}

func TestDBInterleaving(t *testing.T) {
	N := uint64(16)
	d := uint64(8)
	numBytes := uint64(len([]byte("string 16")))

	DBs := make([]*Database, numBytes)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, uint64(1<<10) /* n */, uint64(32) /* log q */)

	for n := uint64(0); n < numBytes; n++ {
		val := make([]uint64, N)
		for i := uint64(0); i < N; i++ {
			arr := []byte("string " + fmt.Sprint(i))
			if uint64(len(arr)) > n {
				val[i] = uint64(arr[n])
			} else {
				val[i] = 0
			}
		}
		DBs[n] = MakeDB(N, d, &p, val)
	}

	D := pir.ConcatDBs(DBs, &p)

	for i := uint64(0); i < N; i++ {
		val := make([]byte, numBytes)
		for n := uint64(0); n < numBytes; n++ {
			val[n] = byte(D.GetElem(i + N*n))
		}
		fmt.Printf("Got '%s' instead of '%s'\n", string(val), "string "+fmt.Sprint(i))
		if strings.TrimRight(string(val), "\x00") != "string "+fmt.Sprint(i) {
			panic("Failure")
		}
	}
}

// Print the BW used by GulliverPIR
func TestGulliverPIRBW(t *testing.T) {
	N := SEC_PARAM
	d := uint64(2048)

	log_N, _ := strconv.Atoi(os.Getenv("LOG_N"))
	D, _ := strconv.Atoi(os.Getenv("D"))
	if log_N != 0 {
		N = uint64(1 << log_N)
	}
	if D != 0 {
		d = uint64(D)
	}

	pir := GulliverPIR{}
	p := pir.PickParams(N, d, SEC_PARAM, LOGQ)
	DB := SetupDB(N, d, &p)

	fmt.Printf("Executing with entries consisting of %d (>= 1) bits; p is %d; packing factor is %d; number of DB elems per entry is %d.\n",
		d, p.P, DB.Info.Packing, DB.Info.Ne)

	pir.GetBW(DB.Info, p)
}

// Test GulliverPIR correctness on DB with short entries.
func TestGulliverPIR(t *testing.T) {
	N := uint64(1 << 20)
	d := uint64(8)
	pir := GulliverPIR{}
	p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

	DB := MakeRandomDB(N, d, &p)
	RunPIR(&pir, DB, p, []uint64{262144})
}
