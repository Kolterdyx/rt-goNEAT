# Getting Started

## Installation

```bash
go get github.com/Kolterdyx/rt-goNEAT/v1
```

## Prerequisites

You need:
1. A **seed genome** — the minimal neural network topology your organisms start with
2. A **configuration** — the NEAT algorithm parameters
3. A **simulation loop** — your code that drives the ALife world

---

## Step 1: Define your seed genome

The seed genome determines the input/output topology every organism starts with. You can load one from a file or build one programmatically.

### Loading from a file (recommended)

Plain-text `.neat` genome files look like this:

```
genomestart 1
trait 1 0.1 0 0 0 0 0 0 0
node 1 0 1 1
node 2 0 1 1
node 3 0 1 3
node 4 0 0 2
gene 1 1 4 1.5 false 1 0 true
gene 2 2 4 2.5 false 2 0 true
gene 3 3 4 3.5 false 3 0 true
genomeend 1
```

See [Genomes](genomes.md) for the full format reference. The library ships with several example genomes in `data/`.

```go
reader, err := genetics.NewGenomeReaderFromFile("data/xorstartgenes")
if err != nil {
    log.Fatal(err)
}
startGenome, err := reader.Read()
if err != nil {
    log.Fatal(err)
}
```

---

## Step 2: Create Options

Options control the NEAT mutation and speciation parameters.

### Inline (for prototyping)

```go
opts := &neat.Options{
    // Speciation
    CompatThreshold:    3.0,
    DisjointCoeff:      1.0,
    ExcessCoeff:        1.0,
    MutdiffCoeff:       0.4,
    GenCompatMethod:    neat.GenomeCompatibilityMethodLinear,

    // Initial population size
    PopSize: 100,

    // Mutation probabilities (one structural mutation per offspring, otherwise weights)
    MutateAddNodeProb:     0.03,
    MutateAddLinkProb:     0.05,
    MutateConnectSensors:  0.10,
    MutateLinkWeightsProb: 0.80,
    MutateOnlyProb:        0.25,

    // Mating
    MateMultipointProb:    0.60,
    MateMultipointAvgProb: 0.30,
    MateSinglepointProb:   0.10,
    MateOnlyProb:          0.20,
    InterspeciesMateRate:  0.001,

    WeightMutPower: 2.5,
    NewLinkTries:   20,

    // Activation functions for new hidden nodes
    NodeActivators:     []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
    NodeActivatorsProb: []float64{1.0},

    LogLevel: "info",
}
```

### From a YAML config file (recommended for production)

```go
opts, err := neat.ReadNeatOptionsFromFile("config/myworld.neat.yml")
```

See [Configuration](configuration.md) for all parameters and example config files.

---

## Step 3: Create and run the simulation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"

    "github.com/Kolterdyx/rt-goNEAT/v1/alife"
    "github.com/Kolterdyx/rt-goNEAT/v1/neat"
    "github.com/Kolterdyx/rt-goNEAT/v1/neat/genetics"
    neatmath "github.com/Kolterdyx/rt-goNEAT/v1/neat/math"
)

func main() {
    // 1. Load seed genome
    reader, err := genetics.NewGenomeReaderFromFile("data/xorstartgenes")
    if err != nil {
        log.Fatal(err)
    }
    startGenome, err := reader.Read()
    if err != nil {
        log.Fatal(err)
    }

    // 2. Configure
    opts := &neat.Options{
        CompatThreshold:       3.0,
        DisjointCoeff:         1.0,
        ExcessCoeff:           1.0,
        MutdiffCoeff:          0.4,
        GenCompatMethod:       neat.GenomeCompatibilityMethodLinear,
        PopSize:               20,
        MutateAddNodeProb:     0.03,
        MutateAddLinkProb:     0.05,
        MutateLinkWeightsProb: 0.80,
        MutateOnlyProb:        0.25,
        MateMultipointProb:    0.60,
        MateMultipointAvgProb: 0.30,
        MateSinglepointProb:   0.10,
        MateOnlyProb:          0.20,
        WeightMutPower:        2.5,
        NewLinkTries:          20,
        NodeActivators:        []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
        NodeActivatorsProb:    []float64{1.0},
    }

    // 3. Create simulation
    sim, err := alife.NewSimulation(context.Background(), startGenome, opts)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Run 1000 ticks
    for tick := 0; tick < 1000; tick++ {
        sim.Step()

        orgs := sim.Organisms()

        // Activate each organism's network with random inputs
        for _, org := range orgs {
            net, err := org.Phenotype()
            if err != nil {
                continue
            }
            // Load inputs (must match the number of input nodes in your seed genome)
            _ = net.LoadSensors([]float64{rand.Float64(), rand.Float64()})
            activated, _ := net.Activate()
            if !activated {
                continue
            }
            outputs := net.ReadOutputs()
            _ = outputs // use in your simulation logic
        }

        // Reproduce: pick two random parents, create an offspring
        if len(orgs) >= 2 {
            p1 := orgs[rand.Intn(len(orgs))]
            p2 := orgs[rand.Intn(len(orgs))]
            _, _ = sim.ReproduceSexual(p1, p2)
        }

        // Kill random organism to keep population bounded
        if len(orgs) > 30 {
            target := orgs[rand.Intn(len(orgs))]
            _ = sim.Kill(target)
        }

        if tick%100 == 0 {
            fmt.Printf("Tick %d: %d organisms, %d species\n",
                tick, len(sim.Organisms()), len(sim.Population.Species))
        }
    }
}
```

---

## What happens under the hood

When you call `NewSimulation`:
1. `PopSize` copies of the seed genome are created, each with slightly perturbed weights
2. All organisms are speciated (assigned to species by compatibility distance)
3. Innovation tracking is initialized

When you call `ReproduceAsexual(parent)`:
1. The parent's genome is duplicated
2. Mutations are applied probabilistically (structural or weight-based)
3. The offspring is speciated and registered in the population
4. `Observer.OnOrganismBorn` fires if any observers are registered

When you call `Kill(org)`:
1. The organism is removed from its species and from `Population.Organisms`
2. Empty species are pruned
3. `Observer.OnOrganismDied` fires

See [Simulation API](simulation-api.md) for complete method details.
