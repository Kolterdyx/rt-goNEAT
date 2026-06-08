# Speciation

Speciation groups organisms by genetic similarity, protecting novel structures from being immediately outcompeted by established ones. In rt-goNEAT, speciation happens automatically and in real-time: every newly reproduced organism is assigned to a species the moment it is added to the population.

---

## How Species Form

When a new organism is created (via reproduction or `Population.AddOrganism`):

1. The library iterates all existing species and computes the **compatibility distance** between the new organism's genome and the **first organism** in each species (the species representative).
2. The species with the smallest compatibility distance below `CompatThreshold` wins — the new organism joins it.
3. If no existing species is compatible, a **new species is created** and the organism becomes its founding member.

```
New organism
    │
    ├── Compatible with species 1 (dist=1.2 < threshold=3.0) ✓
    ├── Compatible with species 2 (dist=2.7 < threshold=3.0) ✓ ← best match
    └── Not compatible with species 3 (dist=4.1 > threshold=3.0)
    
→ Joins species 2
```

---

## Compatibility Distance Formula

```
distance = DisjointCoeff × (disjoint / maxGenes)
         + ExcessCoeff   × (excess  / maxGenes)
         + MutdiffCoeff  × avgWeightDiff
```

Where:
- **disjoint** — genes present in one genome but not the other (in the overlapping innovation range)
- **excess** — genes in the longer genome past the shorter genome's highest innovation number
- **maxGenes** — the larger of the two genome sizes (normalises for genome length)
- **avgWeightDiff** — average absolute weight difference across matching genes

**Calculation method:** `GenCompatMethod`
- `"linear"` — exact O(n) alignment pass, correct for all genome sizes
- `"fast"` — approximate O(log n) method, best for genomes with many genes (>20)

---

## Tuning the Threshold

`CompatThreshold` is the most impactful speciation parameter.

| Threshold | Effect |
|-----------|--------|
| 1.0 – 2.0 | Many small species; strong niche protection; slower convergence |
| 2.5 – 4.0 | Moderate species count; good balance (recommended starting point: 3.0) |
| 5.0+ | Few species; most organisms compete directly; fast but risky for innovation |

To find a good value, log the number of species per tick and adjust:

```go
type SpeciesLogger struct{}

func (s *SpeciesLogger) OnOrganismBorn(_ *alife.Simulation, _ *genetics.Organism) {}
func (s *SpeciesLogger) OnOrganismDied(_ *alife.Simulation, _ *genetics.Organism) {}

func (s *SpeciesLogger) OnSpeciesFormed(sim *alife.Simulation, sp *genetics.Species) {
    fmt.Printf("[tick %d] new species %d (total: %d)\n",
        sim.Tick(), sp.Id, len(sim.Population.Species))
}

func (s *SpeciesLogger) OnSpeciesExtinct(sim *alife.Simulation, sp *genetics.Species) {
    fmt.Printf("[tick %d] species %d extinct\n", sim.Tick(), sp.Id)
}
```

---

## Species Lifecycle

### Formation

A species is created by `createFirstSpecies` when no compatible species exists. The new species:
- Gets an incremented `LastSpecies` ID
- Is marked `IsNovel = true` (protected: its first tick, `Age` is not incremented)
- Contains the founding organism

### Aging

Species age is tracked via `species.Age`. In ALife mode you are responsible for aging species according to your simulation's time model. The library does not auto-age species — there is no epoch cycle to trigger it.

If you want periodic aging (e.g., every N ticks):

```go
if sim.Tick() % 100 == 0 {
    for _, sp := range sim.Population.Species {
        if sp.IsNovel {
            sp.IsNovel = false
        } else {
            sp.Age++
        }
    }
}
```

### Extinction

Species are pruned automatically by `PruneEmptySpecies`, which is called internally by `Kill` and `KillWhere`. A species is removed when it has zero members. The `Observer.OnSpeciesExtinct` callback fires for each pruned species.

---

## Inspecting Species

```go
for _, sp := range sim.Population.Species {
    fmt.Printf("Species %d: age=%d, size=%d, novel=%v\n",
        sp.Id, sp.Age, sp.Size(), sp.IsNovel)

    oldest := sp.FindOldest()
    if oldest != nil {
        fmt.Printf("  oldest member: tick %d\n", oldest.Generation)
    }
}
```

### Finding an organism's species

```go
org := sim.Organisms()[0]
fmt.Println(org.Species.Id, org.Species.Age)
```

---

## Species Diversity Pressure

In traditional NEAT, species compete for offspring via fitness sharing. In ALife mode, the simulation itself is the pressure. Species that produce organisms that survive and reproduce more will naturally grow; species whose organisms die without reproducing will shrink and eventually go extinct.

To deliberately favour diversity, your simulation code can:

1. **Reward exploring organisms** — give more reproduction opportunities to organisms in small or young species
2. **Penalise crowding** — reduce energy of organisms in overpopulated species
3. **Cross-species mating** — occasionally mate organisms from different species (set `InterspeciesMateRate > 0`)

---

## Manual Speciation

For advanced use cases, you can trigger speciation manually:

```go
// Create a genome externally
genome := buildCustomGenome()
org, _ := genetics.NewOrganism(genome, int(sim.Tick()))

// Register with the population (triggers speciation)
newSpecies, err := sim.Population.AddOrganism(sim.Context(), org)
if newSpecies != nil {
    fmt.Printf("created new species %d\n", newSpecies.Id)
}
```

`AddOrganism` returns the newly created species (or `nil` if the organism joined an existing one), and appends the organism to `Population.Organisms`.
