package pir

import (
	"fmt"
	"math"

	_ "embed"
)

type Params struct {
	N       uint64 // LWR secret dimension
	Uniform uint64 // LWR secret distribution

	L uint64 // DB height
	M uint64 // DB width

	LogQ uint64 // (logarithm of) hint modulus
	Logq uint64 // (logarithm of) query modulus
	P    uint64 // plaintext modulus
}

func (p *Params) deltah() float64 {
	Q := uint64(1 << p.LogQ)
	return float64(p.P) / float64(Q)
}
func (p *Params) deltaq() float64 {
	Q := uint64(1 << p.LogQ)
	q := uint64(1 << p.Logq)
	return float64(q) / float64(Q)
}
func (p *Params) deltai() uint64 {
	q := uint64(1 << p.Logq)
	return uint64(float64(q) / float64(p.P))
}
func (p *Params) deltaa() float64 {
	q := uint64(1 << p.Logq)
	return float64(p.P) / float64(q)
}

func ApproxSquareDatabase(d uint64) (uint64, uint64) {
	l := uint64(math.Floor(math.Sqrt(float64(d))))
	m := uint64(math.Ceil(float64(d) / float64(l)))
	return l, m
}

func (p *Params) PrintParams() {
	fmt.Printf("Working with: n=%d; db size=2^%d (l=%d, m=%d); logQ=%d; logq=%d; p=%d; unifrom=%d\n",
		p.N, int(math.Log2(float64(p.L))+math.Log2(float64(p.M))), p.L, p.M, p.LogQ, p.Logq,
		p.P, p.Uniform)
}
