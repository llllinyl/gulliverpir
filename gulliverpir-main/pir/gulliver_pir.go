package pir

// #cgo CFLAGS: -O3 -march=native
// #include "pir.h"
import "C"
import (
	"math"
)

type GulliverPIR struct{}

func (pi *GulliverPIR) Name() string {
	return "GulliverPIR"
}

func (pi *GulliverPIR) PickParams(N, d, n, logq uint64) Params {
	good_p := Params{}
	found := false

	// Iteratively refine p and DB dims, until find tight values
	for mod_p := uint64(2); ; mod_p += 1 {
		m := uint64(math.Sqrt(float64(d)))
		l := m
		p := Params{
			N:    n,
			LogQ: logq,
			L:    l,
			M:    m,
		}
		p.PickParams(false, m)

		if p.P < mod_p {
			if !found {
				panic("Error; should not happen")
			}
			good_p.PrintParams()
			return good_p
		}

		good_p = p
		found = true
	}
}
func (pi *GulliverPIR) Init(info DBinfo, p Params) State {
	A := MatrixRand(p.M, p.N, p.Logq, 0)
	return MakeState(A)
}

func (pi *GulliverPIR) Setup(DB *Database, shared State, p Params) (State, Msg) {
	A := shared.Data[0]
	H := MatrixMul(DB.Data, A)
	return MakeState(), MakeMsg(H)
}

func (pi *GulliverPIR) Query(i uint64, shared State, p Params, info DBinfo) (State, Msg) {
	A := shared.Data[0]

	secret := MatrixRand(p.N, 1, p.Uniform, 0)
	secret.Sub(p.Uniform / 2)
	query := MatrixMul(A, secret)
	for j := uint64(0); j < p.M; j++ {
		query.Data[j] = C.Elem(math.Round(float64(query.Data[j]) * p.deltaq()))
	}
	query.Data[i%p.M] += C.Elem(p.deltai())
	return MakeState(secret), MakeMsg(query)
}

func (pi *GulliverPIR) Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {
	ans := new(Matrix)
	for _, qu := range query.Data {
		a := MatrixMulVec(DB.Data, qu.Data[0])
		ans.Concat(a)
	}
	return MakeMsg(ans)
}

func (pi *GulliverPIR) Recover(i uint64, offline Msg, query Msg, answer Msg,
	shared State, client State, p Params, info DBinfo) uint64 {
	secret := client.Data[0]
	H := offline.Data[0]
	ans := answer.Data[0]
	row := i / p.M
	interm := MatrixMulVec(H, secret)
	interm_r := float64(interm.Data[row]) * p.deltah()
	ans_r := float64(ans.Data[row]) * p.deltaa()

	val_f := ans_r - interm_r
	// fmt.Printf("row 32 interm:%f", interm_r)
	// fmt.Printf("row 32 ans:%f", ans_r)
	ans.MatrixSub(interm)

	val := uint64(math.Round(val_f)) % p.P
	return val
}
