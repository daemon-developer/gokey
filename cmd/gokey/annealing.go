package main

import (
	"math"
)

// SimulatedAnnealing represents the simulated annealing optimizer
type SimulatedAnnealing struct {
	T0 float64
	K  float64
	P0 float64
	N  int
	KN float64
}

// NewSimulatedAnnealing initializes a simulated annealing optimizer
// for keyboard layout optimization.
//
// The annealing process starts with high exploration (accepting many suboptimal moves)
// and gradually transitions to exploitation (refining the best solutions found).
//
// Adjust T0, K, and N to control the balance between exploration and exploitation.
// Higher T0 and lower K/N ratio favor exploration, while lower T0 and higher K/N
// ratio favor exploitation.
//
// It's often beneficial to run multiple annealing processes with different
// parameter sets to find the best results for your specific problem.
func NewSimulatedAnnealing(iterations int) *SimulatedAnnealing {
	sa := &SimulatedAnnealing{
		// Initial temperature: controls initial acceptance probability of worse solutions
		// Higher values (5-20) allow more exploration early on
		// Lower values (1-5) focus more on exploitation from the start
		T0: 15.0,

		// Cooling rate: controls how quickly temperature decreases
		// Higher values (20-100) cool faster, lower values (1-20) cool slower
		// Should be adjusted proportionally with N to maintain cooling profile
		K: float64(iterations) / 2000.0,

		// Probability scaling factor: typically left at 1.0
		// Adjust only if you need to fine-tune acceptance probabilities
		P0: 1.0,

		// Number of iterations: total steps in the annealing process
		// More iterations (100,000-1,000,000) allow for more thorough exploration
		// Fewer iterations (10,000-100,000) for quicker, less thorough searches
		N: iterations,
	}

	// Normalized cooling rate: used in temperature calculation
	// This value should remain constant if K and N are scaled proportionally
	sa.KN = sa.K / float64(sa.N)

	return sa
}

// Temperature calculates the temperature at iteration i
func (sa *SimulatedAnnealing) Temperature(i int) float64 {
	return sa.T0 * math.Exp(-float64(i)*sa.KN)
}

// CutoffP calculates the cutoff probability
func (sa *SimulatedAnnealing) CutoffP(de float64, i int) float64 {
	t := sa.Temperature(i)
	return sa.P0 * math.Exp(-de/t)
}

// AcceptTransition determines whether to accept a transition
func (sa *SimulatedAnnealing) AcceptTransition(de float64, i int) bool {
	if de < 0.0 {
		return true
	}
	pDe := sa.CutoffP(de, i)
	r := r.Float64()
	return r < pDe
}

// GetSimulationRange returns the range for simulation
func (sa *SimulatedAnnealing) GetSimulationRange() (int, int) {
	return 1, sa.N + 1
}
