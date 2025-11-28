package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// Create detectors for 3 neighbors
	neighbor1 := NewPhiAccrualDetector(8.0, 1000, 100*time.Millisecond)
	neighbor2 := NewPhiAccrualDetector(8.0, 1000, 100*time.Millisecond)
	neighbor3 := NewPhiAccrualDetector(8.0, 1000, 100*time.Millisecond)
	neighbours := []*PhiAccrualDetector{neighbor1, neighbor2, neighbor3}

	// Simulate neighbor-1, sends heartbeats every 1 second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			neighbor1.Heartbeat()
			fmt.Printf("[%s][n-1] Heartbeat received\n", time.Now().Format("15:04:05.000"))
		}
	}()

	// Simulate neighbor-2, sends heartbeats every 1 second with ±50ms jitter
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			jitter := time.Duration(rand.Intn(100)-50) * time.Millisecond
			time.Sleep(jitter)
			neighbor2.Heartbeat()
			fmt.Printf("[%s][n-2] Heartbeat received\n", time.Now().Format("15:04:05.000"))
		}
	}()

	// Simulate neighbor-3, sends heartbeats for 5 sec then dies
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		deadline := time.Now().Add(5 * time.Second)
		for t := range ticker.C {
			if t.After(deadline) {
				return
			}
			neighbor3.Heartbeat()
			fmt.Printf("[%s][n-3] Heartbeat received\n", time.Now().Format("15:04:05.000"))
		}
	}()

	// Wait a bit for initial samples to build up
	time.Sleep(3 * time.Second)

	// Check phi values every 500ms
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Format("15:04:05.000")
		for i, neighbour := range neighbours {
			phi := neighbour.Phi()
			status := ""

			if phi > neighbour.threshold {
				status = "(SUSPECTED DOWN)"
			}

			fmt.Printf("[%s][n-%d] phi=%.2f %s\n", now, i+1, phi, status)
		}
		fmt.Println()
	}
}
