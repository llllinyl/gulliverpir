package pir

import (
	"fmt"
	"math"
)

type DBinfo struct {
	Num        uint64 // number of DB entries.
	Row_length uint64 // number of bits per DB entry.

	Packing uint64 // number of DB entries per Z_p elem, if log(p) > DB entry size.
	Ne      uint64 // number of Z_p elems per DB entry, if DB entry size > log(p).

	X uint64 // tunable param that governs communication,
	// must be in range [1, ne] and must be a divisor of ne;
	// represents the number of times the scheme is repeated.
	P    uint64 // plaintext modulus.
	Logq uint64 // (logarithm of) ciphertext modulus.

	// For in-memory DB compression
	Basis     uint64
	Squishing uint64
	Cols      uint64
}

type Database struct {
	Info DBinfo
	Data *Matrix
}

// Find smallest l, m such that l*m >= N*ne and ne divides l, where ne is
// the number of Z_p elements per DB entry determined by row_length and p.
func ApproxSquareDatabaseDims(N, row_length, p uint64) (uint64, uint64) {
	db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
	l := uint64(math.Floor(math.Sqrt(float64(db_elems))))

	rem := l % elems_per_entry
	if rem != 0 {
		l += elems_per_entry - rem
	}

	m := uint64(math.Ceil(float64(db_elems) / float64(l)))

	return l, m
}

// Find smallest l, m such that l*m >= N*ne and ne divides l, where ne is
// the number of Z_p elements per DB entry determined by row_length and p, and m >=
// lower_bound_m.
func ApproxDatabaseDims(N, row_length, p, lower_bound_m uint64) (uint64, uint64) {
	l, m := ApproxSquareDatabaseDims(N, row_length, p)
	if m >= lower_bound_m {
		return l, m
	}

	m = lower_bound_m
	db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
	l = uint64(math.Ceil(float64(db_elems) / float64(m)))

	rem := l % elems_per_entry
	if rem != 0 {
		l += elems_per_entry - rem
	}

	return l, m
}

func SetupDB(Num, row_length uint64, p *Params) *Database {
	if (Num == 0) || (row_length == 0) {
		panic("Empty database!")
	}

	D := new(Database)

	D.Info.Num = Num
	D.Info.Row_length = row_length
	D.Info.P = p.P
	D.Info.Logq = p.LogQ

	db_elems, elems_per_entry, entries_per_elem := Num_DB_entries(Num, row_length, p.P)
	D.Info.Ne = elems_per_entry
	D.Info.X = D.Info.Ne
	D.Info.Packing = entries_per_elem

	for D.Info.Ne%D.Info.X != 0 {
		D.Info.X += 1
	}

	D.Info.Basis = 0
	D.Info.Squishing = 0

	fmt.Printf("Total packed DB size is ~%f MB\n",
		float64(p.L*p.M)*math.Log2(float64(p.P))/(1024.0*1024.0*8.0))

	if db_elems > p.L*p.M {
		panic("Params and database size don't match")
	}

	if p.L%D.Info.Ne != 0 {
		panic("Number of DB elems per entry must divide DB height")
	}

	return D
}

func MakeRandomDB(Num, row_length uint64, p *Params) *Database {
	D := SetupDB(Num, row_length, p)
	D.Data = MatrixRand(p.L, p.M, 0, p.P)

	// Map DB elems to [-p/2; p/2]
	D.Data.Sub(p.P / 2)

	return D
}

func MakeDB(Num, row_length uint64, p *Params, vals []uint64) *Database {
	D := SetupDB(Num, row_length, p)
	D.Data = MatrixZeros(p.L, p.M)

	if uint64(len(vals)) != Num {
		panic("Bad input DB")
	}

	if D.Info.Packing > 0 {
		// Pack multiple DB elems into each Z_p elem
		at := uint64(0)
		cur := uint64(0)
		coeff := uint64(1)
		for i, elem := range vals {
			cur += (elem * coeff)
			coeff *= (1 << row_length)
			if ((i+1)%int(D.Info.Packing) == 0) || (i == len(vals)-1) {
				D.Data.Set(cur, at/p.M, at%p.M)
				at += 1
				cur = 0
				coeff = 1
			}
		}
	} else {
		// Use multiple Z_p elems to represent each DB elem
		for i, elem := range vals {
			for j := uint64(0); j < D.Info.Ne; j++ {
				D.Data.Set(Base_p(D.Info.P, elem, j), (uint64(i)/p.M)*D.Info.Ne+j, uint64(i)%p.M)
			}
		}
	}

	// Map DB elems to [-p/2; p/2]
	D.Data.Sub(p.P / 2)

	return D
}
