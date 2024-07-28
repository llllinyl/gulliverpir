package pir

// #cgo CFLAGS: -O3 -march=native
// #include "pir.h"
import "C"
import (
	"fmt"
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
		l, m := ApproxSquareDatabaseDims(N, d, mod_p)

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

func (pi *GulliverPIR) PickParamsGivenDimensions(l, m, n, logq uint64) Params {
	p := Params{
		N:    n,
		LogQ: logq,
		L:    l,
		M:    m,
	}
	p.PickParams(false, m)
	return p
}

// Works for GulliverPIR because vertical concatenation doesn't increase
// the number of LWE samples (so don't need to change LWE params)
func (pi *GulliverPIR) ConcatDBs(DBs []*Database, p *Params) *Database {
	if len(DBs) == 0 {
		panic("Should not happen")
	}

	if DBs[0].Info.Num != p.L*p.M {
		panic("Not yet implemented")
	}

	rows := DBs[0].Data.Rows
	for j := 1; j < len(DBs); j++ {
		if DBs[j].Data.Rows != rows {
			panic("Bad input")
		}
	}

	D := new(Database)
	D.Data = MatrixZeros(0, 0)
	D.Info = DBs[0].Info
	D.Info.Num *= uint64(len(DBs))
	p.L *= uint64(len(DBs))

	for j := 0; j < len(DBs); j++ {
		D.Data.Concat(DBs[j].Data.SelectRows(0, rows))
	}

	return D
}

func (pi *GulliverPIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.L*p.N*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.M*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.L*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
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
	num_queries := uint64(len(query.Data)) // number of queries in the batch of queries
	batch_sz := DB.Data.Rows / num_queries // how many rows of the database each query in the batch maps to
	q := uint64(1 << p.Logq)
	last := uint64(0)

	// Run GulliverPIR's answer routine for each query in the batch
	for batch, q := range query.Data {
		if batch == int(num_queries-1) {
			batch_sz = DB.Data.Rows - last
		}
		a := MatrixMulVec(DB.Data, q.Data[0])
		ans.Concat(a)
		last += batch_sz
	}
	for j := uint64(0); j < uint64(p.M*num_queries); j++ {
		ans.Data[j] = ans.Data[j] % C.Elem(q)
	}
	return MakeMsg(ans)
}

func (pi *GulliverPIR) Recover(i uint64, offline Msg, query Msg, answer Msg,
	shared State, client State, p Params, info DBinfo) uint64 {
	secret := client.Data[0]
	H := offline.Data[0]
	ans := answer.Data[0]

	row := i / p.M
	interm := MatrixMul(H, secret)
	for j := uint64(0); j < p.M; j++ {
		interm.Data[j] = C.Elem(math.Round(float64(interm.Data[j]) * p.deltah()))
		ans.Data[j] = C.Elem(math.Round(float64(interm.Data[j]) * p.deltaa()))
	}
	ans.MatrixSub(interm)

	val := uint64(ans.Data[row] % C.Elem(p.P))
	return val
}
