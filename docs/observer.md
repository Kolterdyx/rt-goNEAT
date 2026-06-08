# Observer Pattern

The `alife.Observer` interface lets you react to organism and species lifecycle events without polling the organism list every tick.

```go
import "github.com/Kolterdyx/rt-goNEAT/alife"
```

---

## The Interface

```go
type Observer interface {
    OnOrganismBorn(sim *Simulation, org *genetics.Organism)
    OnOrganismDied(sim *Simulation, org *genetics.Organism)
    OnSpeciesFormed(sim *Simulation, species *genetics.Species)
    OnSpeciesExtinct(sim *Simulation, species *genetics.Species)
}
```

All four methods must be implemented. Methods you don't need can be no-ops.

### When each method fires

| Method | Fires when |
|--------|-----------|
| `OnOrganismBorn` | After `ReproduceAsexual` or `ReproduceSexual` successfully adds an organism |
| `OnOrganismDied` | After `Kill` or `KillWhere` removes an organism |
| `OnSpeciesFormed` | When a reproduced organism cannot join any existing species and creates a new one |
| `OnSpeciesExtinct` | When `Kill` or `KillWhere` causes a species to become empty and it is pruned |

**Timing:** all callbacks are synchronous, called from the goroutine that triggered the event. The simulation state is already updated by the time the callback fires — `org.IsAlive()` is `false` in `OnOrganismDied`, and the species is already removed from `Population.Species` in `OnSpeciesExtinct`.

---

## Registering Observers

```go
sim.RegisterObserver(myObserver)
```

Multiple observers can be registered. They are notified in registration order:

```go
sim.RegisterObserver(&Logger{})
sim.RegisterObserver(&MetricsCollector{})
sim.RegisterObserver(&Visualizer{})
```

Register observers before starting the simulation loop to ensure no events are missed.

---

## Example: Logging Observer

```go
type LogObserver struct {
    mu sync.Mutex
    f  *os.File
}

func (l *LogObserver) OnOrganismBorn(sim *alife.Simulation, org *genetics.Organism) {
    l.mu.Lock()
    defer l.mu.Unlock()
    fmt.Fprintf(l.f, "%d BORN  species=%d genome=%d\n",
        sim.Tick(), org.Species.Id, org.Genotype.Id)
}

func (l *LogObserver) OnOrganismDied(sim *alife.Simulation, org *genetics.Organism) {
    age := int(sim.Tick()) - org.Generation
    l.mu.Lock()
    defer l.mu.Unlock()
    fmt.Fprintf(l.f, "%d DIED  species=%d age=%d\n",
        sim.Tick(), org.Species.Id, age)
}

func (l *LogObserver) OnSpeciesFormed(sim *alife.Simulation, sp *genetics.Species) {
    l.mu.Lock()
    defer l.mu.Unlock()
    fmt.Fprintf(l.f, "%d SPECIES_FORMED id=%d\n", sim.Tick(), sp.Id)
}

func (l *LogObserver) OnSpeciesExtinct(sim *alife.Simulation, sp *genetics.Species) {
    l.mu.Lock()
    defer l.mu.Unlock()
    fmt.Fprintf(l.f, "%d SPECIES_EXTINCT id=%d age=%d\n",
        sim.Tick(), sp.Id, sp.Age)
}
```

---

## Example: Statistics Collector

```go
type Stats struct {
    mu          sync.Mutex
    born        int
    died        int
    speciesEver int
    peakPop     int
}

func (s *Stats) OnOrganismBorn(sim *alife.Simulation, _ *genetics.Organism) {
    s.mu.Lock()
    s.born++
    if n := len(sim.Organisms()); n > s.peakPop {
        s.peakPop = n
    }
    s.mu.Unlock()
}

func (s *Stats) OnOrganismDied(_ *alife.Simulation, _ *genetics.Organism) {
    s.mu.Lock()
    s.died++
    s.mu.Unlock()
}

func (s *Stats) OnSpeciesFormed(_ *alife.Simulation, _ *genetics.Species) {
    s.mu.Lock()
    s.speciesEver++
    s.mu.Unlock()
}

func (s *Stats) OnSpeciesExtinct(_ *alife.Simulation, _ *genetics.Species) {}

func (s *Stats) Print() {
    s.mu.Lock()
    defer s.mu.Unlock()
    fmt.Printf("born=%d died=%d species_ever=%d peak_pop=%d\n",
        s.born, s.died, s.speciesEver, s.peakPop)
}
```

---

## Example: Async Dispatch

If your observer does heavy work (writing to a database, updating a UI), dispatch to a background goroutine to avoid blocking the simulation:

```go
type AsyncObserver struct {
    events chan Event
}

type Event struct {
    kind string
    tick int64
    org  *genetics.Organism
    sp   *genetics.Species
}

func NewAsyncObserver() *AsyncObserver {
    obs := &AsyncObserver{events: make(chan Event, 1000)}
    go obs.process()
    return obs
}

func (a *AsyncObserver) process() {
    for ev := range a.events {
        switch ev.kind {
        case "born":
            saveToDatabase(ev.tick, ev.org)
        case "died":
            updateDashboard(ev.tick, ev.org)
        }
    }
}

func (a *AsyncObserver) OnOrganismBorn(sim *alife.Simulation, org *genetics.Organism) {
    // non-blocking send; drop event if buffer full
    select {
    case a.events <- Event{kind: "born", tick: sim.Tick(), org: org}:
    default:
    }
}

func (a *AsyncObserver) OnOrganismDied(sim *alife.Simulation, org *genetics.Organism) {
    select {
    case a.events <- Event{kind: "died", tick: sim.Tick(), org: org}:
    default:
    }
}

func (a *AsyncObserver) OnSpeciesFormed(_ *alife.Simulation, _ *genetics.Species) {}
func (a *AsyncObserver) OnSpeciesExtinct(_ *alife.Simulation, _ *genetics.Species) {}
```

---

## Minimal No-Op Implementation

For testing or when you only care about a subset of events:

```go
type NoOpObserver struct{}

func (n *NoOpObserver) OnOrganismBorn(_ *alife.Simulation, _ *genetics.Organism)  {}
func (n *NoOpObserver) OnOrganismDied(_ *alife.Simulation, _ *genetics.Organism)  {}
func (n *NoOpObserver) OnSpeciesFormed(_ *alife.Simulation, _ *genetics.Species)  {}
func (n *NoOpObserver) OnSpeciesExtinct(_ *alife.Simulation, _ *genetics.Species) {}
```

Embed `NoOpObserver` in your observer to only override what you need:

```go
type BirthTracker struct {
    NoOpObserver  // embed for default no-op implementations
    count int
}

func (b *BirthTracker) OnOrganismBorn(_ *alife.Simulation, _ *genetics.Organism) {
    b.count++
}
```
