package pir

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed params.csv
var lwr_params string

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
	Q := 1 << p.LogQ
	return float64(p.P) / float64(Q)
}
func (p *Params) deltaq() float64 {
	Q := 1 << p.LogQ
	q := 1 << p.Logq
	return float64(q) / float64(Q)
}
func (p *Params) deltai() uint64 {
	q := 1 << p.Logq
	return uint64(float64(q) / float64(p.P))
}
func (p *Params) deltaa() float64 {
	q := 1 << p.Logq
	return float64(p.P) / float64(q)
}

func (p *Params) PickParams(doublepir bool, samples ...uint64) {
	if p.N == 0 || p.Logq == 0 {
		panic("Need to specify n and q!")
	}

	num_samples := uint64(0)
	for _, ns := range samples {
		if ns > num_samples {
			num_samples = ns
		}
	}

	lines := strings.Split(lwr_params, "\n")
	for _, l := range lines[1:] {
		line := strings.Split(l, ",")
		logn, _ := strconv.ParseUint(line[0], 10, 64)
		logm, _ := strconv.ParseUint(line[1], 10, 64)
		logQ, _ := strconv.ParseUint(line[2], 10, 64)
		logq, _ := strconv.ParseUint(line[3], 10, 64)

		if (p.N == uint64(1<<logn)) &&
			(num_samples <= uint64(1<<logm)) &&
			(p.LogQ == uint64(logQ)) &&
			(p.Logq == uint64(logq)) {

			uniform, _ := strconv.ParseFloat(line[4], 64)
			p.Uniform = uint64(uniform)
			mod, _ := strconv.ParseUint(line[5], 10, 64)
			p.P = uint64(1 << mod)

			if p.Uniform == 0.0 || p.P == 0 {
				panic("Params invalid!")
			}
			return
		}
	}

	fmt.Printf("Searched for %d, %d-by-%d, %d,\n", p.N, p.L, p.M, p.Logq)
	panic("No suitable params known!")
}

func (p *Params) PrintParams() {
	fmt.Printf("Working with: n=%d; db size=2^%d (l=%d, m=%d); logQ=%d; logq=%d; p=%d; unifrom=%d\n",
		p.N, int(math.Log2(float64(p.L))+math.Log2(float64(p.M))), p.L, p.M, p.LogQ, p.Logq,
		p.P, p.Uniform)
}
