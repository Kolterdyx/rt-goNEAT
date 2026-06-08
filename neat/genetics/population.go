package genetics

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/Kolterdyx/rt-goNEAT/v4/neat"
	"github.com/pkg/errors"
)

// Population is a group of Organisms and the species they belong to.
type Population struct {
	// Species within the population
	Species []*Species
	// All organisms in the population (master list; kept in sync with Species membership)
	Organisms []*Organism
	// LastSpecies is the highest species ID assigned so far
	LastSpecies int

	// tracker owns innovation tracking and node-ID generation
	tracker *InnovationTracker

	// mu guards Species and Organisms from concurrent birth/death events
	mu sync.RWMutex
}

// NewPopulation constructs a population seeded from a single template genome.
func NewPopulation(g *Genome, opts *neat.Options) (*Population, error) {
	if opts.PopSize <= 0 {
		return nil, fmt.Errorf("wrong population size in the context: %d", opts.PopSize)
	}
	pop := newPopulation()
	if err := pop.spawn(g, opts); err != nil {
		return nil, err
	}
	return pop, nil
}

// NewPopulationRandom constructs a population of random topologies.
// See NewGenomeRand for parameter details.
func NewPopulationRandom(in, out, maxHidden int, recurrent bool, linkProb float64, opts *neat.Options) (*Population, error) {
	if opts.PopSize <= 0 {
		return nil, fmt.Errorf("wrong population size in the context: %d", opts.PopSize)
	}
	pop := newPopulation()
	for count := 0; count < opts.PopSize; count++ {
		gen, err := newGenomeRand(count, in, out, rand.Intn(maxHidden), maxHidden, recurrent, linkProb, opts)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create random population")
		}
		org, err := NewOrganism(gen, 1)
		if err != nil {
			return nil, err
		}
		pop.Organisms = append(pop.Organisms, org)
	}
	pop.tracker.nextNodeId = int32(in + out + maxHidden + 1)
	pop.tracker.nextInnovNum = int64((in+out+maxHidden)*(in+out+maxHidden) + 1)

	if err := pop.speciate(opts.NeatContext(), pop.Organisms); err != nil {
		return nil, err
	}
	return pop, nil
}

// Verify runs a consistency check on all genomes in the population (debugging aid).
func (p *Population) Verify() (bool, error) {
	res := true
	var err error
	for _, o := range p.Organisms {
		res, err = o.Genotype.verify()
		if err != nil {
			return false, err
		}
	}
	return res, nil
}

// AddOrganism speciates the organism and registers it in the population. Thread-safe.
// Returns the newly created species if one was formed for this organism, or nil if it
// joined an existing species.
func (p *Population) AddOrganism(ctx context.Context, org *Organism) (*Species, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	speciesBefore := len(p.Species)
	if err := p.speciate(ctx, []*Organism{org}); err != nil {
		return nil, err
	}
	p.Organisms = append(p.Organisms, org)
	org.alive = true
	if len(p.Species) > speciesBefore {
		return p.Species[len(p.Species)-1], nil
	}
	return nil, nil
}

// RemoveOrganism removes an organism from the population and its species. Thread-safe.
func (p *Population) RemoveOrganism(org *Organism) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if org.Species != nil {
		if _, err := org.Species.removeOrganism(org); err != nil {
			return err
		}
	}
	orgs := make([]*Organism, 0, len(p.Organisms))
	for _, o := range p.Organisms {
		if o != org {
			orgs = append(orgs, o)
		}
	}
	p.Organisms = orgs
	org.alive = false
	return nil
}

// PruneEmptySpecies removes any species that currently have no member organisms.
func (p *Population) PruneEmptySpecies() []*Species {
	p.mu.Lock()
	defer p.mu.Unlock()
	alive := make([]*Species, 0, len(p.Species))
	var pruned []*Species
	for _, s := range p.Species {
		if len(s.Organisms) > 0 {
			alive = append(alive, s)
		} else {
			pruned = append(pruned, s)
		}
	}
	p.Species = alive
	return pruned
}

// ClearInnovations resets the innovation map. Call this if you want to allow
// the same structural change to get a fresh innovation number going forward.
func (p *Population) ClearInnovations() {
	p.tracker.clear()
}

// --- InnovationsObserver implementation (used by genome mutation operators) ---

func (p *Population) StoreInnovation(innovation Innovation) {
	p.tracker.StoreInnovation(innovation)
}

func (p *Population) Innovations() []Innovation {
	return p.tracker.Innovations()
}

func (p *Population) NextInnovationNumber() int64 {
	return p.tracker.NextInnovationNumber()
}

func (p *Population) FindLinkInnovation(inNodeId, outNodeId int, isRecurrent bool) *Innovation {
	return p.tracker.FindLinkInnovation(inNodeId, outNodeId, isRecurrent)
}

func (p *Population) FindNodeInnovation(inNodeId, outNodeId int, oldInnovNum int64) *Innovation {
	return p.tracker.FindNodeInnovation(inNodeId, outNodeId, oldInnovNum)
}

// --- NodeIdGenerator implementation (used by genome mutation operators) ---

func (p *Population) NextNodeId() int {
	return p.tracker.NextNodeId()
}

// Snapshot returns a shallow copy of the current organism list, safe for iteration after
// the call returns regardless of concurrent birth/death events.
func (p *Population) Snapshot() []*Organism {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Organism, len(p.Organisms))
	copy(out, p.Organisms)
	return out
}

// newPopulation is the private constructor.
func newPopulation() *Population {
	return &Population{
		Species:   make([]*Species, 0),
		Organisms: make([]*Organism, 0),
		tracker:   newInnovationTracker(),
	}
}

// spawn fills a new population from template genome g, introducing small weight perturbations.
func (p *Population) spawn(g *Genome, opts *neat.Options) (err error) {
	for count := 0; count < opts.PopSize; count++ {
		newGenome, err := g.duplicate(count)
		if err != nil {
			return err
		}
		if _, err = newGenome.mutateLinkWeights(1.0, 1.0, gaussianMutator); err != nil {
			return err
		}
		newOrganism, err := NewOrganism(newGenome, 1)
		if err != nil {
			return err
		}
		p.Organisms = append(p.Organisms, newOrganism)
	}
	nextNodeId, err := g.getLastNodeId()
	if err != nil {
		return err
	}
	p.tracker.nextNodeId = int32(nextNodeId + 1)

	nextInnovNum, err := g.getNextGeneInnovNum()
	if err != nil {
		return err
	}
	p.tracker.nextInnovNum = nextInnovNum - 1

	return p.speciate(opts.NeatContext(), p.Organisms)
}

// speciate assigns each organism to the most compatible existing species, or creates a new
// one if none are compatible. Must be called with mu held (write lock) or during construction.
func (p *Population) speciate(ctx context.Context, organisms []*Organism) error {
	if len(organisms) == 0 {
		return errors.New("no organisms to speciate from")
	}
	opts, found := neat.FromContext(ctx)
	if !found {
		return neat.ErrNEATOptionsNotFound
	}

	for _, currOrg := range organisms {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if len(p.Species) == 0 {
			createFirstSpecies(p, currOrg)
		} else {
			if opts.CompatThreshold == 0 {
				return errors.New("compatibility threshold is set to ZERO - will not find any compatible species")
			}
			var bestCompatible *Species
			bestCompatValue := math.MaxFloat64
			for _, currSpecies := range p.Species {
				compOrg := currSpecies.firstOrganism()
				if compOrg != nil {
					currCompat := currOrg.Genotype.compatibility(compOrg.Genotype, opts)
					if currCompat < opts.CompatThreshold && currCompat < bestCompatValue {
						bestCompatible = currSpecies
						bestCompatValue = currCompat
					}
				}
			}
			if bestCompatible != nil {
				if neat.LogLevel == neat.LogLevelDebug {
					neat.DebugLog(fmt.Sprintf("POPULATION: Compatible species [%d] found for baby organism [%d]",
						bestCompatible.Id, currOrg.Genotype.Id))
				}
				bestCompatible.addOrganism(currOrg)
				currOrg.Species = bestCompatible
			} else {
				createFirstSpecies(p, currOrg)
			}
		}
	}
	return nil
}
