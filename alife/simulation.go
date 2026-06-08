package alife

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/Kolterdyx/rt-goNEAT/neat"
	"github.com/Kolterdyx/rt-goNEAT/neat/genetics"
)

// Simulation is the top-level ALife controller. It wraps a genetics.Population and
// exposes a real-time API: explicit Reproduce / Kill calls, a monotonic tick counter,
// and optional Observer callbacks.
type Simulation struct {
	// Population is the underlying genetics population. Direct access is allowed for
	// read-only introspection; mutations must go through Simulation methods.
	Population *genetics.Population

	ctx       context.Context
	tick      atomic.Int64
	mu        sync.RWMutex // protects observers slice only
	observers []Observer
}

// NewSimulation seeds a population from startGenome and returns a Simulation.
// opts must contain at least CompatThreshold, PopSize, and mutation probability fields.
func NewSimulation(ctx context.Context, startGenome *genetics.Genome, opts *neat.Options) (*Simulation, error) {
	pop, err := genetics.NewPopulation(startGenome, opts)
	if err != nil {
		return nil, err
	}
	return &Simulation{
		Population: pop,
		ctx:        neat.NewContext(ctx, opts),
	}, nil
}

// RegisterObserver adds an observer. Safe to call before the simulation loop starts.
func (s *Simulation) RegisterObserver(o Observer) {
	s.mu.Lock()
	s.observers = append(s.observers, o)
	s.mu.Unlock()
}

// Tick returns the current simulation tick.
func (s *Simulation) Tick() int64 {
	return s.tick.Load()
}

// Step advances the tick counter by 1.
func (s *Simulation) Step() {
	s.tick.Add(1)
}

// Organisms returns a snapshot of all current organisms. The slice is safe to iterate
// after the call returns; it is not a live view of the population.
func (s *Simulation) Organisms() []*genetics.Organism {
	return s.Population.Snapshot()
}

// Context returns the context carrying NEAT options.
func (s *Simulation) Context() context.Context {
	return s.ctx
}

func (s *Simulation) notifyBorn(org *genetics.Organism, newSpecies *genetics.Species) {
	s.mu.RLock()
	obs := s.observers
	s.mu.RUnlock()
	for _, o := range obs {
		o.OnOrganismBorn(s, org)
	}
	if newSpecies != nil {
		for _, o := range obs {
			o.OnSpeciesFormed(s, newSpecies)
		}
	}
}

func (s *Simulation) notifyDied(org *genetics.Organism, extinctSpecies []*genetics.Species) {
	s.mu.RLock()
	obs := s.observers
	s.mu.RUnlock()
	for _, o := range obs {
		o.OnOrganismDied(s, org)
	}
	for _, sp := range extinctSpecies {
		for _, o := range obs {
			o.OnSpeciesExtinct(s, sp)
		}
	}
}
