package pir

// #cgo CFLAGS: -O3 -march=native
// #include "pir.h"
import "C"
import (
	"math"
)

// GulliverPIR represents the Gulliver Private Information Retrieval scheme.
type GulliverPIR struct{}

// Name returns the name of the PIR scheme.
func (pi *GulliverPIR) Name() string {
	return "GulliverPIR"
}

// PickParams iteratively refines parameters for the PIR scheme until suitable values are found.
func (pi *GulliverPIR) PickParams(N, d, n, logQ, logq uint64) Params {
	l, m := ApproxSquareDatabase(d)
	Delta := logQ - logq
	p := Params{
		N:       n,
		LogQ:    logQ,
		L:       l,
		M:       m,
		Logq:    logq,
		Uniform: uint64(1 << Delta),
	}
	k := float64(math.Log2(float64(d)))
	t := (float64(logq) - k/2 + 1) / 2
	p.P = uint64(1 << uint64(math.Floor(t)))
	p.PrintParams()
	return p
}

// Init initializes the state for the PIR scheme.
func (pi *GulliverPIR) Init(info DBinfo, p Params) State {
	A := MatrixRand(p.M, p.N, p.LogQ, 0)
	return MakeState(A)
}

// Setup prepares the database and shared state for the PIR scheme.
func (pi *GulliverPIR) Setup(DB *Database, shared State, p Params) (State, Msg) {
	A := shared.Data[0]
	H := MatrixMul(DB.Data, A)
	DB.Data.Add(p.P / 2)
	DB.Squish()
	return MakeState(), MakeMsg(H)
}

// Query generates a query for the specified index using the shared state.
func (pi *GulliverPIR) Query(i uint64, shared State, p Params, info DBinfo) (State, Msg) {
	A := shared.Data[0]
	secret := MatrixRand(p.N, 1, p.Uniform, 0)
	secret.Sub(p.Uniform / 2)
	query := MatrixMul(A, secret)

	// Apply scaling and rounding to each element of the query.
	for j := uint64(0); j < p.M; j++ {
		query.Data[j] = C.Elem(math.Round(float64(query.Data[j]) * p.deltaq()))
	}
	query.Data[i%p.M] += C.Elem(p.deltai())

	// Ensure the query dimensions match the compressed database.
	if p.M%info.Squishing != 0 {
		query.AppendZeros(info.Squishing - (p.M % info.Squishing))
	}

	return MakeState(secret), MakeMsg(query)
}

// Answer generates the server's response to a batch of queries.
func (pi *GulliverPIR) Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {
	ans := new(Matrix)
	numQueries := uint64(len(query.Data))
	batchSize := DB.Data.Rows / numQueries

	var last uint64
	for batch, q := range query.Data {
		if batch == int(numQueries-1) {
			batchSize = DB.Data.Rows - last
		}
		a := MatrixMulVecPacked(DB.Data.SelectRows(last, batchSize),
			q.Data[0],
			DB.Info.Basis,
			DB.Info.Squishing)
		ans.Concat(a)
		last += batchSize
	}
	return MakeMsg(ans)
}

// Recover reconstructs the original database element from the query and answer.
func (pi *GulliverPIR) Recover(i uint64, batchIndex uint64, offline Msg, query Msg, answer Msg,
	shared State, client State, p Params, info DBinfo) uint64 {
	secret := client.Data[0]
	H := offline.Data[0]
	ans := answer.Data[0]
	row := i / p.M

	// Calculate the offset for the query element.
	ratio := p.P / 2
	var offset uint64
	for j := uint64(0); j < p.M; j++ {
		offset += ratio * query.Data[0].Get(j, 0)
	}
	offset %= (1 << p.Logq)
	offset = (1 << p.Logq) - offset

	interm := MatrixMul(H, secret)
	var vals []uint64
	for j := row * info.Ne; j < (row+1)*info.Ne; j++ {
		item0 := float64(interm.Data[j]) * p.deltah()
		item1 := float64(ans.Data[j]+C.Elem(offset)) * p.deltaa()
		denoised := uint64(math.Round(item1-item0)) % p.P
		vals = append(vals, denoised)
	}
	return ReconstructElem(vals, i, info)
}

// Reset resets the database to its original state.
func (pi *GulliverPIR) Reset(DB *Database, p Params) {
	DB.Unsquish()
	DB.Data.Sub(p.P / 2)
}
