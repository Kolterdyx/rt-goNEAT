// Package alife provides a real-time Artificial Life simulation controller built on top of
// the rt-goNEAT genetics engine. It replaces the epoch-based Experiment framework with
// explicit, simulation-driven organism birth and death.
package alife

import (
	"github.com/Kolterdyx/rt-goNEAT/v4/neat/genetics"
)

// Observer receives lifecycle events from a Simulation.
// All methods are called synchronously from the Simulation goroutine; implementations
// that do expensive work should dispatch to their own goroutine.
type Observer interface {
	// OnOrganismBorn is called after a new organism has been added to the population.
	OnOrganismBorn(sim *Simulation, org *genetics.Organism)
	// OnOrganismDied is called after an organism has been removed from the population.
	OnOrganismDied(sim *Simulation, org *genetics.Organism)
	// OnSpeciesFormed is called when a reproduced organism could not join an existing
	// species and therefore created a new one.
	OnSpeciesFormed(sim *Simulation, species *genetics.Species)
	// OnSpeciesExtinct is called for each species removed during PruneEmptySpecies.
	OnSpeciesExtinct(sim *Simulation, species *genetics.Species)
}
