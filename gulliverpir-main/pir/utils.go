package pir

import (
	"fmt"
	"math"
)

type State struct {
	Data []*Matrix
}

type CompressedState struct {
	Seed *PRGKey
}

type Msg struct {
	Data []*Matrix
}

func (m *Msg) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

type MsgSlice struct {
	Data []Msg
}

func (m *MsgSlice) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

func MakeState(elems ...*Matrix) State {
	return State{
		Data: append(make([]*Matrix, 0), elems...),
	}
}
func MakeCompressedState(elem *PRGKey) CompressedState {
	st := CompressedState{}
	st.Seed = elem
	return st
}

func MakeMsg(elems ...*Matrix) Msg {
	return Msg{
		Data: append(make([]*Matrix, 0), elems...),
	}
}

func MakeMsgSlice(elems ...Msg) MsgSlice {
	return MsgSlice{
		Data: append(make([]Msg, 0), elems...),
	}
}

// Returns the i-th elem in the representation of m in base p.
func Base_p(p, m, i uint64) uint64 {
	for j := uint64(0); j < i; j++ {
		m = m / p
	}
	return (m % p)
}

// Reconstruct_from_base_p reconstructs an element from its base-p decomposition.
func Reconstruct_from_base_p(p uint64, vals []uint64) uint64 {
	var result uint64
	var coefficient uint64 = 1

	for _, value := range vals {
		result += value * coefficient
		coefficient *= p
	}

	return result
}

// Returns how many entries in Z_p are needed to represent an element in Z_q
func Compute_num_entries_base_p(p, logQ uint64) uint64 {
	logP := math.Log2(float64(p))
	return uint64(math.Ceil(float64(logQ) / logP))
}

// Num_DB_entries calculates the number of Z_p elements required to represent a database
// with N entries, each of 'rowlength' bits, using a prime number p.
func Num_DB_entries(N, rowLength, p uint64) (dbEntries, numElemsPerDBEntry, entriesPerZpElem uint64) {
	if float64(rowLength) <= math.Log2(float64(p)) {
		// If the row length is less than or equal to the log2 of p,
		// then pack multiple DB entries into a single Z_p element.
		logP := uint64(math.Log2(float64(p)))
		entriesPerZpElem = logP / rowLength
		dbEntries = uint64(math.Ceil(float64(N) / float64(entriesPerZpElem)))

		if dbEntries == 0 || dbEntries > N {
			fmt.Printf("Calculated number of entries is incorrect: %d for N = %d\n", dbEntries, N)
			panic("Invalid number of database entries")
		}

		return dbEntries, 1, entriesPerZpElem
	}

	// If the row length is greater than log2 of p, use multiple Z_p elements
	// to represent a single DB entry.
	numElemsPerDBEntry = Compute_num_entries_base_p(p, rowLength)
	return N * numElemsPerDBEntry, numElemsPerDBEntry, 0
}
