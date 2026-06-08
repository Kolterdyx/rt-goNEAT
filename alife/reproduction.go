package alife

import (
	"github.com/Kolterdyx/rt-goNEAT/v1/neat/genetics"
)

// ReproduceAsexual creates a new organism by cloning and mutating parent.
// The offspring is speciated and added to the population automatically.
// Returns the new organism and any error.
func (s *Simulation) ReproduceAsexual(parent *genetics.Organism) (*genetics.Organism, error) {
	offspring, err := s.Population.CreateMutatedOffspring(s.ctx, parent)
	if err != nil {
		return nil, err
	}
	offspring.Generation = int(s.Tick())

	newSpecies, err := s.Population.AddOrganism(s.ctx, offspring)
	if err != nil {
		return nil, err
	}

	s.notifyBorn(offspring, newSpecies)
	return offspring, nil
}

// ReproduceSexual creates a new organism by crossing parent1 and parent2.
// parent1 is the primary parent: its disjoint/excess genes are preferred in the crossover.
// The offspring is speciated and added to the population automatically.
// Returns the new organism and any error.
func (s *Simulation) ReproduceSexual(parent1, parent2 *genetics.Organism) (*genetics.Organism, error) {
	offspring, err := s.Population.CreateMatedOffspring(s.ctx, parent1, parent2)
	if err != nil {
		return nil, err
	}
	offspring.Generation = int(s.Tick())

	newSpecies, err := s.Population.AddOrganism(s.ctx, offspring)
	if err != nil {
		return nil, err
	}

	s.notifyBorn(offspring, newSpecies)
	return offspring, nil
}
