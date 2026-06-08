package alife

import (
	"github.com/Kolterdyx/rt-goNEAT/neat/genetics"
)

// Kill removes a single organism from the simulation immediately.
// Any species that become empty as a result are pruned, and observers are notified.
func (s *Simulation) Kill(org *genetics.Organism) error {
	if err := s.Population.RemoveOrganism(org); err != nil {
		return err
	}
	extinct := s.Population.PruneEmptySpecies()
	s.notifyDied(org, extinct)
	return nil
}

// KillWhere removes every organism for which predicate returns true.
// Returns the count of organisms killed and the first error encountered (if any).
func (s *Simulation) KillWhere(predicate func(*genetics.Organism) bool) (int, error) {
	targets := make([]*genetics.Organism, 0)
	for _, org := range s.Population.Snapshot() {
		if predicate(org) {
			targets = append(targets, org)
		}
	}
	var lastErr error
	for _, org := range targets {
		if err := s.Kill(org); err != nil {
			lastErr = err
		}
	}
	return len(targets), lastErr
}
