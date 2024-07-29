package pir

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
	// "math"
)

// Defines the interface for PIR with preprocessing schemes
type PIR interface {
	Name() string

	Init(info DBinfo, p Params) State

	Setup(DB *Database, shared State, p Params) (State, Msg)

	Query(i uint64, shared State, p Params, info DBinfo) (State, Msg)

	Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg

	Recover(i uint64, offline Msg, query Msg, answer Msg, shared State, client State, p Params, info DBinfo) uint64
}

// Run full GulliverPIR scheme (offline + online phases) single query.
func RunPIR(pi PIR, DB *Database, p Params, i uint64) (float64, float64) {
	fmt.Printf("Executing %s\n", pi.Name())
	//fmt.Printf("Memory limit: %d\n", debug.SetMemoryLimit(math.MaxInt64))
	debug.SetGCPercent(-1)
	bw := float64(0)

	shared_state := pi.Init(DB.Info, p)

	fmt.Println("Setup...")
	start := time.Now()
	server_state, offline_download := pi.Setup(DB, shared_state, p)
	printTime(start)
	comm := float64(offline_download.Size() * uint64(p.LogQ) / (8.0 * 1024.0))
	fmt.Printf("\t\tOffline download: %f KB\n", comm)
	bw += comm
	runtime.GC()

	fmt.Println("Building query...")
	start = time.Now()
	var client_state []State
	var query MsgSlice
	cs, qu := pi.Query(i, shared_state, p, DB.Info)
	client_state = append(client_state, cs)
	query.Data = append(query.Data, qu)
	runtime.GC()
	printTime(start)
	comm = float64(query.Size() * uint64(p.Logq) / (8.0 * 1024.0))
	fmt.Printf("\t\tOnline upload: %f KB\n", comm)
	bw += comm
	runtime.GC()

	fmt.Println("Answering query...")
	start = time.Now()
	answer := pi.Answer(DB, query, server_state, shared_state, p)
	elapsed := printTime(start)
	rate := printRate(p, elapsed, 1)
	comm = float64(answer.Size() * uint64(p.Logq) / (8.0 * 1024.0))
	fmt.Printf("\t\tOnline download: %f KB\n", comm)
	bw += comm
	runtime.GC()

	fmt.Println("Reconstructing...")
	start = time.Now()
	val := pi.Recover(i, offline_download,
		query.Data[0], answer, shared_state,
		client_state[0], p, DB.Info)
	realnum := DB.Data.Get(i/p.M, i%p.M) % p.P
	if realnum != val {
		fmt.Printf("querying index %d --: Got %d instead of %d\n",
			i, val, realnum)
		panic("Reconstruct failed!")
	}
	fmt.Printf("Get index %d : %d \n", i, realnum)
	fmt.Println("Success!")
	printTime(start)

	runtime.GC()
	debug.SetGCPercent(100)
	return rate, bw
}
