package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"runtime"
)

// ---------------- Ticket Lock ----------------
type TicketLock struct {
	ticket int32
	turn   int32
}

func (l *TicketLock) Lock() {
	myturn := atomic.AddInt32(&l.ticket, 1) - 1
	for atomic.LoadInt32(&l.turn) != myturn {
		runtime.Gosched() // yield to other goroutines
	}
}

func (l *TicketLock) Unlock() {
	atomic.AddInt32(&l.turn, 1)
}

// ---------------- CAS Lock ----------------
type CASLock struct {
	flag int32
}

func (l *CASLock) Lock() {
	for !atomic.CompareAndSwapInt32(&l.flag, 0, 1) {
		runtime.Gosched() // spin until successful
	}
}

func (l *CASLock) Unlock() {
	atomic.StoreInt32(&l.flag, 0)
}

// ---------------- Benchmark ----------------
func benchmarkLock(lock interface {
	Lock()
	Unlock()
}, goroutines, iterations int) time.Duration {

	var totalWait int64
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				start := time.Now()
				lock.Lock()
				wait := time.Since(start)
				atomic.AddInt64(&totalWait, wait.Nanoseconds())

				// critical section (simulate some work)
				time.Sleep(time.Microsecond)

				lock.Unlock()
			}
		}()
	}

	wg.Wait()
	avgWait := totalWait / int64(goroutines*iterations)
	return time.Duration(avgWait)
}

// ---------------- Main ----------------
func main() {
	const iterations = 1000
	goroutineCounts := []int{2, 5, 10, 20, 50}

	fmt.Printf("%-12s %-12s %-12s\n", "Goroutines", "TicketLock", "CASLock")
	for _, g := range goroutineCounts {
		ticket := &TicketLock{}
		cas := &CASLock{}

		avgTicket := benchmarkLock(ticket, g, iterations)
		avgCAS := benchmarkLock(cas, g, iterations)

		fmt.Printf("%-12d %-12v %-12v\n", g, avgTicket, avgCAS)
	}
}
