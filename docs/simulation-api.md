# Simulation API

The `alife.Simulation` struct is the primary entry point for all ALife interactions. It wraps a `genetics.Population` and provides a thread-safe, event-driven interface.

```go
import "github.com/Kolterdyx/rt-goNEAT/v1/alife"
```

---

## Creating a Simulation

```go
func NewSimulation(
    ctx context.Context,
    startGenome *genetics.Genome,
    opts *neat.Options,
) (*Simulation, error)
```

`NewSimulation` creates a population of `opts.PopSize` organisms, each derived from the seed genome with perturbed weights, speciates them all, and returns the simulation.

**The returned `Simulation` has:**
- `sim.Population` — the underlying `*genetics.Population` (see [Population](#direct-population-access))
- `sim.Tick()` → `0`
- `sim.Organisms()` → `opts.PopSize` organisms

**Error conditions:**
- `opts.PopSize <= 0`
- Genome genesis fails (malformed genome)
- `opts.CompatThreshold == 0` (speciation cannot work)

---

## Tick Counter

The simulation provides a monotonic tick counter for tracking simulation time. It has no effect on organism behavior — it is purely informational and is stamped onto newly born organisms.

```go
func (s *Simulation) Step()       // increment tick by 1
func (s *Simulation) Tick() int64 // read current tick
```

```go
for {
    sim.Step()
    tick := sim.Tick()
    // ... rest of loop
}
```

Newly reproduced organisms have their `Generation` field set to the current tick at the time of reproduction.

---

## Querying Organisms

```go
func (s *Simulation) Organisms() []*genetics.Organism
```

Returns a **snapshot** of all living organisms. The returned slice is a copy — it is safe to iterate over even while reproduction and death are happening concurrently. The slice is not automatically updated; call `Organisms()` again on the next tick.

```go
for _, org := range sim.Organisms() {
    // org.IsAlive() is true here, but may become false
    // on the next tick if Kill was called
    fmt.Println(org.Generation, org.Species.Id)
}
```

For concurrent access patterns, also see [Population.Snapshot](#direct-population-access).

---

## Reproduction

### Asexual (clone + mutate)

```go
func (s *Simulation) ReproduceAsexual(parent *genetics.Organism) (*genetics.Organism, error)
```

Creates a new organism by duplicating the parent's genome and applying probabilistic mutations. The mutation decision tree (checked in order):

1. If `rand < MutateAddNodeProb` → structural: add hidden node
2. Else if `rand < MutateAddLinkProb` → structural: add connection
3. Else if `rand < MutateConnectSensors` → structural: connect a disconnected sensor
4. If none of the above → non-structural: mutate link weights, traits, etc.

The offspring is speciated and added to the population automatically. Its `Generation` is set to `sim.Tick()`.

```go
parent := sim.Organisms()[0]
offspring, err := sim.ReproduceAsexual(parent)
if err != nil {
    log.Println("reproduction failed:", err)
    return
}
fmt.Printf("Born organism %v in species %d\n", offspring, offspring.Species.Id)
```

### Sexual (crossover)

```go
func (s *Simulation) ReproduceSexual(parent1, parent2 *genetics.Organism) (*genetics.Organism, error)
```

Creates a new organism by crossing two parent genomes. `parent1` is the **primary parent**: when parents carry genes the other doesn't (disjoint or excess), the primary parent's genes are inherited.

Crossover type is chosen by probability:
- `MateMultipointProb` → multipoint (random pick at each matching gene)
- `MateMultipointAvgProb / (MateMultipointAvgProb + MateSinglepointProb)` → multipoint-average
- otherwise → single-point

After crossover, a mutation step is applied unless a coin flip lands below `MateOnlyProb` (and the parents are not identical).

```go
orgs := sim.Organisms()
if len(orgs) < 2 {
    return
}
p1 := orgs[0]
p2 := orgs[rand.Intn(len(orgs))]
offspring, err := sim.ReproduceSexual(p1, p2)
```

**Choosing parents:** parent1 should be the "dominant" parent — typically the one you want to contribute more of its structure. In an energy-based simulation, parent1 might be the organism with more energy.

---

## Killing Organisms

### Kill a specific organism

```go
func (s *Simulation) Kill(org *genetics.Organism) error
```

Removes `org` from the population immediately. If the organism's species becomes empty, the species is pruned. Observers are notified after the removal.

```go
err := sim.Kill(org)
if err != nil {
    log.Println("kill failed:", err)
}
// org.IsAlive() is now false
```

**Error conditions:** the organism is not in the population (already killed, or from a different simulation).

### Kill by predicate

```go
func (s *Simulation) KillWhere(predicate func(*genetics.Organism) bool) (int, error)
```

Removes all organisms for which `predicate` returns `true`. Returns the count killed and the first error encountered (processing continues on error).

```go
// Kill all organisms older than 200 ticks
killed, err := sim.KillWhere(func(org *genetics.Organism) bool {
    age := int(sim.Tick()) - org.Generation
    return age > 200
})
fmt.Printf("killed %d old organisms\n", killed)
```

```go
// Kill organisms with depleted energy (using Data field)
type State struct{ Energy float64 }

killed, _ := sim.KillWhere(func(org *genetics.Organism) bool {
    if org.Data == nil {
        return false
    }
    return org.Data.Value.(*State).Energy <= 0
})
```

---

## Observer Registration

```go
func (s *Simulation) RegisterObserver(o Observer)
```

Registers an observer for lifecycle events. Multiple observers can be registered; all are notified in registration order. Observers are called synchronously on the goroutine that triggered the event.

```go
type MyObserver struct{}

func (m *MyObserver) OnOrganismBorn(sim *alife.Simulation, org *genetics.Organism) {
    fmt.Printf("[tick %d] born organism in species %d\n", sim.Tick(), org.Species.Id)
}
func (m *MyObserver) OnOrganismDied(sim *alife.Simulation, org *genetics.Organism)   { /* ... */ }
func (m *MyObserver) OnSpeciesFormed(sim *alife.Simulation, sp *genetics.Species)     { /* ... */ }
func (m *MyObserver) OnSpeciesExtinct(sim *alife.Simulation, sp *genetics.Species)    { /* ... */ }

sim.RegisterObserver(&MyObserver{})
```

See [Observer Pattern](observer.md) for detailed usage.

---

## Accessing the Context

```go
func (s *Simulation) Context() context.Context
```

Returns the `context.Context` that carries the `neat.Options`. You can use this context when calling lower-level population methods directly:

```go
// Manual reproduction without going through Simulation
baby, err := sim.Population.CreateMutatedOffspring(sim.Context(), parent)
if err != nil {
    return err
}
newSpecies, err := sim.Population.AddOrganism(sim.Context(), baby)
```

---

## Direct Population Access

`sim.Population` is a `*genetics.Population` and is accessible for inspection. The following Population methods are safe to call directly:

| Method | Description |
|--------|-------------|
| `Population.Snapshot()` | Thread-safe copy of organism list |
| `Population.Species` | Direct access to species slice (read-only) |
| `Population.LastSpecies` | Highest species ID assigned |
| `Population.Verify()` | Debugging: check all genomes for structural validity |
| `Population.Write(w)` | Serialise all genomes |
| `Population.WriteBySpecies(w)` | Serialise grouped by species |
| `Population.ClearInnovations()` | Reset innovation map |

Do **not** modify `Population.Organisms` or `Population.Species` directly — use `AddOrganism` / `RemoveOrganism` to maintain internal consistency.

---

## Thread Safety

The `Simulation` methods are **not** internally synchronized with each other — the expectation is that your simulation loop runs on one goroutine. Within a single goroutine, all operations are safe.

If you need concurrent access (e.g., parallel network activation while a background goroutine kills organisms), use these patterns:

```go
// Safe: read-only snapshot for parallel activation
orgs := sim.Organisms()  // returns a copy
var wg sync.WaitGroup
for _, org := range orgs {
    wg.Add(1)
    go func(o *genetics.Organism) {
        defer wg.Done()
        net, _ := o.Phenotype()
        net.LoadSensors(inputs)
        net.Activate()
    }(org)
}
wg.Wait()

// Safe: kill/reproduce on one goroutine after parallel activation completes
for _, org := range orgs {
    if shouldDie(org) {
        sim.Kill(org)
    }
}
```

`Population.AddOrganism` and `Population.RemoveOrganism` each acquire a write lock internally, so they are safe to call from multiple goroutines if needed.
