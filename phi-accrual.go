package main

import (
	"math"
	"sync"
	"time"
)

type PhiAccrualDetector struct {
	mu              sync.RWMutex
	threshold       float64 // If phi > threshold, the node is dead
	maxSampleSize   int     // How many heartbeats to remember
	minStdDeviation time.Duration
	intervals       []time.Duration // Circular buffer of past intervals
	lastHeartbeat   time.Time       // When did we last hear from this node?
	stats           Stats
}

func NewPhiAccrualDetector(threshold float64, maxSampleSize int, minStdDeviation time.Duration) *PhiAccrualDetector {
	return &PhiAccrualDetector{
		threshold:       threshold,
		maxSampleSize:   maxSampleSize,
		minStdDeviation: minStdDeviation,
		intervals:       make([]time.Duration, 0, maxSampleSize),
		stats:           Stats{},
	}
}

// When a heartbeat arrives:
// 1. Calculate the interval since the last heartbeat
// 2. Store that interval
// 3. Update the "last seen" timestamp
func (pa *PhiAccrualDetector) HeartbeatAt(timestamp time.Time) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	// First heartbeat - just set the baseline
	if pa.lastHeartbeat.IsZero() {
		pa.lastHeartbeat = timestamp
		return
	}

	diff := timestamp.Sub(pa.lastHeartbeat)
	if len(pa.intervals) >= pa.maxSampleSize {
		pa.stats.Remove(float64(pa.intervals[0]))
		pa.intervals = pa.intervals[1:]
	}

	pa.lastHeartbeat = timestamp
	pa.intervals = append(pa.intervals, diff)
	pa.stats.Add(float64(diff))
}

func (d *PhiAccrualDetector) Heartbeat() {
	d.HeartbeatAt(time.Now())
}

// Assume the heartbeat intervals follow a normal distribution
// phi value is the cumulative distribution function (CDF) of the normal distribution
// phi = -log10(1 - CDF(timeSinceLastHeartbeat))
func (pa *PhiAccrualDetector) Phi() float64 {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if pa.lastHeartbeat.IsZero() {
		return 0
	}

	timeSinceLastHeartbeat := time.Since(pa.lastHeartbeat)
	cdf := pa.stats.CDF(float64(timeSinceLastHeartbeat), float64(pa.minStdDeviation))

	pDown := 1.0 - cdf
	if pDown < 1e-10 {
		pDown = 1e-10 // This gives phi ≈ 10
	}

	return -math.Log10(pDown)
}

func (pa *PhiAccrualDetector) IsAlive() bool {
	return pa.Phi() < pa.threshold
}
