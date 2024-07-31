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

	PickParams(N, d, n, logQ, logq uint64) Params

	Init(info DBinfo, p Params) State

	Setup(DB *Database, shared State, p Params) (State, Msg)

	Query(i uint64, shared State, p Params, info DBinfo) (State, Msg)

	Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg

	Recover(i uint64, batch_index uint64, offline Msg, query Msg, answer Msg, shared State, client State, p Params, info DBinfo) uint64

	Reset(DB *Database, p Params)
}

// RunPIR executes the full GulliverPIR scheme for a single query,
// which includes both offline and online phases.
func RunPIR(pi PIR, DB *Database, p Params, queryIndex uint64) (float64, float64) {
	fmt.Printf("Executing %s\n", pi.Name())
	debug.SetGCPercent(-1)
	var bw float64
	var clientState []State
	var query MsgSlice

	// Initialize the shared state.
	sharedState := pi.Init(DB.Info, p)

	// Perform the setup phase.
	fmt.Println("Setup...")
	startTime := time.Now()
	serverState, offlineDownload := pi.Setup(DB, sharedState, p)
	printTime(startTime)
	communicationSize := calculateCommunicationSize(offlineDownload.Size(), p.LogQ)
	fmt.Printf("\tOffline download: %f KB\n", communicationSize)
	bw += communicationSize
	runtime.GC()

	// Build the query for the given index.
	fmt.Println("Building query...")
	startTime = time.Now()
	cs, qu := pi.Query(queryIndex, sharedState, p, DB.Info)
	clientState = append(clientState, cs)
	query.Data = append(query.Data, qu)
	printTime(startTime)
	communicationSize = calculateCommunicationSize(query.Size(), p.Logq)
	fmt.Printf("\tOnline upload: %f KB\n", communicationSize)
	bw += communicationSize
	runtime.GC()

	// Answer the query.
	fmt.Println("Answering query...")
	startTime = time.Now()
	answer := pi.Answer(DB, query, serverState, sharedState, p)
	elapsedTime := printTime(startTime)
	transferRate := printRate(p, elapsedTime, 1)
	communicationSize = calculateCommunicationSize(answer.Size(), p.Logq)
	fmt.Printf("\tOnline download: %f KB\n", communicationSize)
	bw += communicationSize
	runtime.GC()

	// Reset the database to its original state.
	pi.Reset(DB, p)

	// Reconstruct the queried element and verify correctness.
	fmt.Println("Reconstructing...")
	startTime = time.Now()
	reconstructedValue := pi.Recover(queryIndex, 1, offlineDownload,
		query.Data[0], answer, sharedState, clientState[0], p, DB.Info)
	expectedValue := DB.GetElem(queryIndex)
	if reconstructedValue != expectedValue {
		fmt.Printf("querying index %d --: Got %d instead of %d\n", queryIndex, reconstructedValue, expectedValue)
		panic("Reconstruct failed!")
	}
	fmt.Printf("Get index %d : %d \n", queryIndex, expectedValue)
	fmt.Println("Success!")
	printTime(startTime)

	// Restore the garbage collection to its default settings.
	runtime.GC()
	debug.SetGCPercent(100)

	return transferRate, bw
}
