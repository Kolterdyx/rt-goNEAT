package genetics

import (
	"fmt"
	"github.com/Kolterdyx/rt-goNEAT/v4/neat"
	"io"
	"sort"
)

// Species is a group of genetically compatible Organisms.
type Species struct {
	// Unique identifier
	Id int
	// Number of simulation ticks this species has existed
	Age int
	// IsNovel marks a newly created species; novel species skip the first age increment
	IsNovel bool
	// Organisms belonging to this species
	Organisms Organisms
	// IsChecked is a scratch flag used during speciation search
	IsChecked bool
}

// NewSpecies constructs a species with the given ID.
func NewSpecies(id int) *Species {
	return newSpecies(id)
}

// NewSpeciesNovel creates a species that will not age on its first tick (protects new species).
func NewSpeciesNovel(id int, novel bool) *Species {
	s := newSpecies(id)
	s.IsNovel = novel
	return s
}

func newSpecies(id int) *Species {
	return &Species{
		Id:        id,
		Age:       1,
		Organisms: make([]*Organism, 0),
	}
}

// Write serialises the species and all its organism genomes to w.
func (s *Species) Write(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "/* Species #%d : (Size %d) (Age %d) */\n",
		s.Id, len(s.Organisms), s.Age); err != nil {
		return err
	}

	sorted := make(Organisms, len(s.Organisms))
	copy(sorted, s.Organisms)
	sort.Sort(sort.Reverse(sorted))

	for _, org := range sorted {
		if _, err := fmt.Fprintf(w, "/* Organism #%d (tick %d) */\n",
			org.Genotype.Id, org.Generation); err != nil {
			return err
		}
		if err := org.Genotype.Write(w); err != nil {
			return err
		}
	}
	return nil
}

// addOrganism appends an organism to this species.
func (s *Species) addOrganism(o *Organism) {
	s.Organisms = append(s.Organisms, o)
}

// removeOrganism removes an organism from this species and returns whether it was found.
func (s *Species) removeOrganism(org *Organism) (bool, error) {
	orgs := make([]*Organism, 0, len(s.Organisms))
	for _, o := range s.Organisms {
		if o != org {
			orgs = append(orgs, o)
		}
	}
	if len(orgs) != len(s.Organisms)-1 {
		return false, fmt.Errorf("attempt to remove nonexistent Organism from Species with #of organisms: %d", len(s.Organisms))
	}
	s.Organisms = orgs
	return true, nil
}

// firstOrganism returns the first organism in the species, or nil if empty.
func (s *Species) firstOrganism() *Organism {
	if len(s.Organisms) > 0 {
		return s.Organisms[0]
	}
	return nil
}

// Size returns the number of organisms in this species.
func (s *Species) Size() int {
	return len(s.Organisms)
}

// FindOldest returns the organism with the lowest birth tick (longest-lived), or nil if empty.
func (s *Species) FindOldest() *Organism {
	var oldest *Organism
	for _, org := range s.Organisms {
		if oldest == nil || org.Generation < oldest.Generation {
			oldest = org
		}
	}
	return oldest
}

func (s *Species) String() string {
	str := fmt.Sprintf("Species #%d, age=%d, size=%d\n", s.Id, s.Age, len(s.Organisms))
	for _, o := range s.Organisms {
		str += fmt.Sprintf("\t%s\n", o)
	}
	return str
}

// createFirstSpecies creates the very first species in a population and assigns baby to it.
func createFirstSpecies(pop *Population, baby *Organism) {
	if neat.LogLevel == neat.LogLevelDebug {
		neat.DebugLog(fmt.Sprintf("SPECIES: Create first species for baby organism [%d]", baby.Genotype.Id))
	}

	pop.LastSpecies++
	species := NewSpeciesNovel(pop.LastSpecies, true)
	pop.Species = append(pop.Species, species)
	species.addOrganism(baby)
	baby.Species = species

	if neat.LogLevel == neat.LogLevelDebug {
		neat.DebugLog(fmt.Sprintf("SPECIES: # of species in population: %d, new species id: %d",
			len(pop.Species), species.Id))
	}
}
