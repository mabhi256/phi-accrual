package main

import "math"

// Welford's algorithm for online calculation of mean and variance
type Stats struct {
	Mean  float64
	M2    float64 // Σ(x - mean)²
	Count int     // variance = m2 / count
}

func (s *Stats) Add(newValue float64) {
	s.Count++
	delta := newValue - s.Mean         // distance from current mean
	s.Mean += delta / float64(s.Count) // shift mean towards new value
	delta2 := newValue - s.Mean        // distance from new mean
	s.M2 += delta * delta2
}

func (s *Stats) Remove(oldValue float64) {
	if s.Count <= 0 {
		panic("list is empty")
	}

	delta := oldValue - s.Mean
	s.Mean -= delta / float64(s.Count)
	delta2 := oldValue - s.Mean
	s.M2 -= delta * delta2
	s.Count--

	// Ensure M2 doesn't go negative due to floating point errors
	if s.M2 < 0 {
		s.M2 = 0
	}
}

// How far do values deviate from the mean on an average
func (s *Stats) Variance() float64 {
	if s.Count == 0 {
		return 0
	}

	return s.M2 / float64(s.Count) // Population variance
}

// How many standard deviations away from the mean is this value.
func (s *Stats) ZScore(value, minStdDev float64) float64 {
	if s.Count == 0 {
		return 0
	}

	stdDev := math.Sqrt(s.Variance())

	if stdDev < minStdDev {
		stdDev = minStdDev
	}

	return (value - s.Mean) / stdDev
}

// Probability that a random value in our distribution is <= this value.
func (s *Stats) CDF(value, minStdDev float64) float64 {
	z := s.ZScore(value, minStdDev)
	return 0.5 * (1.0 + math.Erf(z/math.Sqrt(2.0)))
}
