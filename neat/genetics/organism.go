package genetics

import (
	"bytes"
	"fmt"
	"github.com/Kolterdyx/rt-goNEAT/v4/neat/network"
)

// Organisms represents a list of organisms
type Organisms []*Organism

// OrganismData is the object to associate implementation specific data with particular organism for various algorithm implementations
type OrganismData struct {
	// The implementation specific data object to be associated with organism
	Value interface{}
}

// Organism is a Genotype (Genome) and Phenotype (Network) pair managed by the simulation.
// Fitness, error, and winner semantics are intentionally absent; the embedding ALife simulation
// is responsible for tracking any performance metrics it needs via the Data field.
type Organism struct {
	// The Organism's genotype
	Genotype *Genome
	// The Species of the Organism
	Species *Species

	// Simulation tick at which this organism was born (repurposes the old generation counter)
	Generation int

	// The utility data transfer object to be used by different simulation implementations.
	Data *OrganismData

	// The flag to be used as utility value
	Flag int

	// The Organism's phenotype
	orgPhenotype *network.Network

	// alive is false after the organism has been removed from the population
	alive bool
}

// NewOrganism creates a new organism with the given genome, born at the given tick.
func NewOrganism(g *Genome, tick int) (org *Organism, err error) {
	org = &Organism{
		Genotype:     g,
		orgPhenotype: g.Phenotype,
		Generation:   tick,
		alive:        true,
	}
	return org, nil
}

// IsAlive reports whether the organism is still part of the population.
func (o *Organism) IsAlive() bool {
	return o.alive
}

// Phenotype returns the phenotype of this organism, building it lazily if needed.
func (o *Organism) Phenotype() (*network.Network, error) {
	if o.orgPhenotype == nil {
		phenotype, err := o.Genotype.Genesis(o.Genotype.Id)
		if err != nil {
			return nil, err
		}
		o.orgPhenotype = phenotype
	}
	return o.orgPhenotype, nil
}

// UpdatePhenotype regenerates the underlying network graph based on a change in the genotype.
func (o *Organism) UpdatePhenotype() (err error) {
	o.orgPhenotype = nil
	o.orgPhenotype, err = o.Genotype.Genesis(o.Genotype.Id)
	return err
}

// MarshalBinary encodes this organism for wired transmission during parallel simulation.
func (o *Organism) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := fmt.Fprintln(&buf, o.Generation, o.Genotype.Id); err != nil {
		return nil, err
	}
	if err := o.Genotype.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary decodes an organism received over the wire.
func (o *Organism) UnmarshalBinary(data []byte) (err error) {
	b := bytes.NewBuffer(data)
	var genotypeId int
	if _, err = fmt.Fscanln(b, &o.Generation, &genotypeId); err != nil {
		return err
	}
	if o.Genotype, err = ReadGenome(b, genotypeId); err != nil {
		return err
	}
	o.alive = true
	return nil
}

func (o *Organism) String() string {
	return fmt.Sprintf("[Organism tick: %d, species: %v, alive: %v]",
		o.Generation, o.Species, o.alive)
}

// Dump returns all organism fields as a string (for debugging).
func (o *Organism) Dump() string {
	b := bytes.NewBufferString("Organism:")
	_, _ = fmt.Fprintln(b, "Generation:", o.Generation)
	_, _ = fmt.Fprintln(b, "Alive:", o.alive)
	_, _ = fmt.Fprintln(b, "Phenotype:", o.orgPhenotype)
	_, _ = fmt.Fprintln(b, "Genotype:", o.Genotype)
	_, _ = fmt.Fprintln(b, "Species:", o.Species)
	_, _ = fmt.Fprintln(b, "Data:", o.Data)
	return b.String()
}

// Organisms sort interface — ordered by birth tick (ascending)

func (f Organisms) Len() int      { return len(f) }
func (f Organisms) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f Organisms) Less(i, j int) bool {
	return f[i].Generation < f[j].Generation
}
