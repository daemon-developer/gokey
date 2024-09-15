package main

import (
	"math"
	"math/rand"
)

// SimulatedAnnealing represents the simulated annealing optimizer
type SimulatedAnnealing struct {
	T0 float64
	K  float64
	P0 float64
	N  int
	KN float64
}

// NewSimulatedAnnealing creates a new SimulatedAnnealing instance with default parameters
func NewSimulatedAnnealing() *SimulatedAnnealing {
	sa := &SimulatedAnnealing{
		T0: 1.5,
		K:  10.0,
		P0: 1.0,
		N:  1000, //15000,
	}
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
	r := rand.Float64()
	return r < pDe
}

// GetSimulationRange returns the range for simulation
func (sa *SimulatedAnnealing) GetSimulationRange() (int, int) {
	return 1, sa.N + 1
}
