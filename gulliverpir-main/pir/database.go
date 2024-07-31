package pir

import (
	"fmt"
	"math"
)

// DBinfo stores metadata about the database structure and parameters.
type DBinfo struct {
	Num        uint64 // Total number of entries in the database.
	Row_length uint64 // Number of bits per database entry.

	Packing uint64 // Number of DB entries per Z_p element for compression.
	Ne      uint64 // Number of Z_p elements per DB entry for expansion.

	X    uint64 // Tunable parameter for communication efficiency.
	P    uint64 // Plaintext modulus.
	Logq uint64 // Logarithm of the ciphertext modulus.

	// Parameters for in-memory database compression.
	Basis     uint64
	Squishing uint64
	Cols      uint64
}

type Database struct {
	Info DBinfo
	Data *Matrix
}

func (DB *Database) Squish() {
	DB.Info.Basis = 10
	DB.Info.Squishing = 3
	DB.Info.Cols = DB.Data.Cols
	DB.Data.Squish(DB.Info.Basis, DB.Info.Squishing)

	// Ensure the parameters are suitable for compression.
	if DB.Info.P > (1<<DB.Info.Basis) || DB.Info.Logq < DB.Info.Basis*DB.Info.Squishing {
		panic("Invalid parameters for compression")
	}
}

func (DB *Database) Unsquish() {
	DB.Data.Unsquish(DB.Info.Basis, DB.Info.Squishing, DB.Info.Cols)
}

// ReconstructElem reconstructs an element from its Z_p representation.
func ReconstructElem(vals []uint64, index uint64, info DBinfo) uint64 {
	q := uint64(1 << info.Logq)
	for i, v := range vals {
		vals[i] = ((v + info.P/2) % q) % info.P
	}
	val := Reconstruct_from_base_p(info.P, vals)

	if info.Packing > 0 {
		val = Base_p((1 << uint(info.Row_length)), val, index%info.Packing)
	}
	return val
}

// GetElem retrieves an element from the database by its index.
func (DB *Database) GetElem(i uint64) uint64 {
	if i >= DB.Info.Num {
		panic("Index out of range")
	}
	col := i % DB.Data.Cols
	row := i / DB.Data.Cols

	if DB.Info.Packing > 0 {
		newI := i / DB.Info.Packing
		col = newI % DB.Data.Cols
		row = newI / DB.Data.Cols
	}

	var vals []uint64
	for j := row * DB.Info.Ne; j < (row+1)*DB.Info.Ne; j++ {
		vals = append(vals, DB.Data.Get(j, col))
	}
	return ReconstructElem(vals, i, DB.Info)
}

// SetupDB initializes a new database with the given parameters.
func SetupDB(Num, row_length uint64, p *Params) *Database {
	if Num == 0 || row_length == 0 {
		panic("Empty database")
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

	fmt.Printf("Total packed DB size is ~%f MB\n", float64(p.L*p.M)*math.Log2(float64(p.P))/(1024.0*1024.0*8.0))

	if db_elems > p.L*p.M {
		panic("Parameters and database size do not match")
	}
	if p.L%D.Info.Ne != 0 {
		panic("Number of DB elements per entry must divide the database height")
	}
	return D
}

// MakeRandomDB creates a new database with random entries.
func MakeRandomDB(Num, row_length uint64, p *Params) *Database {
	D := SetupDB(Num, row_length, p)
	D.Data = MatrixRand(p.L, p.M, 0, p.P)
	D.Data.Sub(p.P / 2)
	return D
}

// MakeDB creates a new database with specified entries.
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

	D.Data.Sub(p.P / 2)
	return D
}
