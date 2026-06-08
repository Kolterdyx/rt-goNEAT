package genetics

import (
	"context"
	"math/rand"

	"github.com/Kolterdyx/rt-goNEAT/v1/neat"
)

// CreateMutatedOffspring creates a new organism by duplicating parent's genome and applying
// probabilistic mutations. The organism is NOT added to the population; call AddOrganism to
// register it.
func (p *Population) CreateMutatedOffspring(ctx context.Context, parent *Organism) (*Organism, error) {
	opts, found := neat.FromContext(ctx)
	if !found {
		return nil, neat.ErrNEATOptionsNotFound
	}

	p.mu.Lock()
	newId := len(p.Organisms)
	p.mu.Unlock()

	newGenome, err := parent.Genotype.duplicate(newId)
	if err != nil {
		return nil, err
	}

	mutStructural := false

	if rand.Float64() < opts.MutateAddNodeProb {
		if _, err = newGenome.mutateAddNode(p, p, opts); err != nil {
			return nil, err
		}
		mutStructural = true
	} else if rand.Float64() < opts.MutateAddLinkProb {
		if _, err = newGenome.mutateAddLink(p, parent.Generation, opts); err != nil {
			return nil, err
		}
		mutStructural = true
	} else if rand.Float64() < opts.MutateConnectSensors {
		if added, err := newGenome.mutateConnectSensors(p, opts); err != nil {
			return nil, err
		} else {
			mutStructural = added
		}
	}

	if !mutStructural {
		if _, err = newGenome.mutateAllNonstructural(opts); err != nil {
			return nil, err
		}
	}

	return NewOrganism(newGenome, parent.Generation+1)
}

// CreateMatedOffspring creates a new organism by crossing parent1 and parent2.
// parent1 is treated as the primary parent: its disjoint and excess genes are preferred
// when the parents differ. An optional mutation step is applied based on opts.MateOnlyProb.
// The organism is NOT added to the population; call AddOrganism to register it.
func (p *Population) CreateMatedOffspring(ctx context.Context, parent1, parent2 *Organism) (*Organism, error) {
	opts, found := neat.FromContext(ctx)
	if !found {
		return nil, neat.ErrNEATOptionsNotFound
	}

	p.mu.Lock()
	newId := len(p.Organisms)
	p.mu.Unlock()

	// Use fitness=1/0 convention so parent1's excess/disjoint genes dominate.
	var newGenome *Genome
	var err error
	switch {
	case rand.Float64() < opts.MateMultipointProb:
		newGenome, err = parent1.Genotype.mateMultipoint(parent2.Genotype, newId, 1.0, 0.0)
	case rand.Float64() < opts.MateMultipointAvgProb/(opts.MateMultipointAvgProb+opts.MateSinglepointProb):
		newGenome, err = parent1.Genotype.mateMultipointAvg(parent2.Genotype, newId, 1.0, 0.0)
	default:
		newGenome, err = parent1.Genotype.mateSinglePoint(parent2.Genotype, newId)
	}
	if err != nil {
		return nil, err
	}

	// Mutate unless mating-only probability prevents it, or parents are identical
	needsMutation := rand.Float64() > opts.MateOnlyProb ||
		parent1.Genotype.Id == parent2.Genotype.Id ||
		parent1.Genotype.compatibility(parent2.Genotype, opts) == 0.0

	if needsMutation {
		mutStructural := false
		if rand.Float64() < opts.MutateAddNodeProb {
			if _, err = newGenome.mutateAddNode(p, p, opts); err != nil {
				return nil, err
			}
			mutStructural = true
		} else if rand.Float64() < opts.MutateAddLinkProb {
			tick := parent1.Generation
			if parent2.Generation > tick {
				tick = parent2.Generation
			}
			if _, err = newGenome.mutateAddLink(p, tick, opts); err != nil {
				return nil, err
			}
			mutStructural = true
		} else if rand.Float64() < opts.MutateConnectSensors {
			if mutStructural, err = newGenome.mutateConnectSensors(p, opts); err != nil {
				return nil, err
			}
		}
		if !mutStructural {
			if _, err = newGenome.mutateAllNonstructural(opts); err != nil {
				return nil, err
			}
		}
	}

	tick := parent1.Generation
	if parent2.Generation > tick {
		tick = parent2.Generation
	}
	return NewOrganism(newGenome, tick+1)
}
