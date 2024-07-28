package pir

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// Defines the interface for PIR with preprocessing schemes
type PIR interface {
	Name() string

	PickParams(N, d, n, logQ uint64) Params
	PickParamsGivenDimensions(l, m, n, logQ uint64) Params

	GetBW(info DBinfo, p Params)

	Init(info DBinfo, p Params) State
	InitCompressed(info DBinfo, p Params) (State, CompressedState)
	DecompressState(info DBinfo, p Params, comp CompressedState) State

	Setup(DB *Database, shared State, p Params) (State, Msg)
	// FakeSetup(DB *Database, p Params) (State, float64) // used for benchmarking online phase

	Query(i uint64, shared State, p Params, info DBinfo) (State, Msg)

	Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg

	Recover(i uint64, batch_index uint64, offline Msg, query Msg, answer Msg, shared State, client State,
		p Params, info DBinfo) uint64

	Reset(DB *Database, p Params) // reset DB to its correct state, if modified during execution
}

// Run full GulliverPIR scheme (offline + online phases).
func RunPIR(pi PIR, DB *Database, p Params, i []uint64) (float64, float64) {
	fmt.Printf("Executing %s\n", pi.Name())
	//fmt.Printf("Memory limit: %d\n", debug.SetMemoryLimit(math.MaxInt64))
	debug.SetGCPercent(-1)

	num_queries := uint64(len(i))
	if DB.Data.Rows/num_queries < DB.Info.Ne {
		panic("Too many queries to handle!")
	}
	batch_sz := DB.Data.Rows / (DB.Info.Ne * num_queries) * DB.Data.Cols
	bw := float64(0)

	shared_state := pi.Init(DB.Info, p)

	fmt.Println("Setup...")
	start := time.Now()
	server_state, offline_download := pi.Setup(DB, shared_state, p)
	printTime(start)
	comm := float64(offline_download.Size() * uint64(p.Logq) / (8.0 * 1024.0))
	fmt.Printf("\t\tOffline download: %f KB\n", comm)
	bw += comm
	runtime.GC()

	fmt.Println("Building query...")
	start = time.Now()
	var client_state []State
	var query MsgSlice
	for index, _ := range i {
		index_to_query := i[index] + uint64(index)*batch_sz
		cs, qu := pi.Query(index_to_query, shared_state, p, DB.Info)
		client_state = append(client_state, cs)
		query.Data = append(query.Data, qu)
	}
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
	rate := printRate(p, elapsed, len(i))
	comm = float64(answer.Size() * uint64(p.Logq) / (8.0 * 1024.0))
	fmt.Printf("\t\tOnline download: %f KB\n", comm)
	bw += comm
	runtime.GC()

	fmt.Println("Reconstructing...")
	start = time.Now()
	for index, _ := range i {
		index_to_query := i[index] + uint64(index)*batch_sz
		val := pi.Recover(index_to_query, uint64(index), offline_download,
			query.Data[index], answer, shared_state,
			client_state[index], p, DB.Info)

		if DB.GetElem(index_to_query) != val {
			fmt.Printf("Batch %d (querying index %d -- row should be >= %d): Got %d instead of %d\n",
				index, index_to_query, DB.Data.Rows/4, val, DB.GetElem(index_to_query))
			panic("Reconstruct failed!")
		}
	}
	fmt.Println("Success!")
	printTime(start)

	runtime.GC()
	debug.SetGCPercent(100)
	return rate, bw
}
