# Core Concepts

## What is NEAT?

NEAT (NeuroEvolution of Augmenting Topologies) is an algorithm for evolving artificial neural networks. It is distinctive because it evolves both the **weights** of connections and the **topology** (structure) of the network simultaneously, starting from minimal networks and growing complexity only when useful.

rt-goNEAT adapts NEAT for **Artificial Life** simulations, where the evolutionary pressures come from the simulation environment rather than a fitness function.

---

## Genotype and Phenotype

Every organism in rt-goNEAT has two representations:

```
Genome (genotype)  →  Network (phenotype)
     ↑                       ↑
Blueprint           Actual neural network
Encoded as genes    Ready for activation
```

The **Genome** stores the blueprint: a list of nodes (neurons) and genes (connections between them), each tagged with an innovation number. The **Network** is built from the genome and is what actually computes outputs from inputs. Building the network from the genome is called **genesis**.

You work with the phenotype for computation and the genotype for inheritance.

### Accessing the Phenotype

```go
org := sim.Organisms()[0]
net, err := org.Phenotype()  // builds the network lazily on first call
if err != nil {
    // handle error
}
// net is *network.Network, ready to activate
```

If you mutate the genome directly, call `org.UpdatePhenotype()` to rebuild the network.

---

## Organisms

An `Organism` is a genotype/phenotype pair. In ALife mode, organisms carry no fitness score — the simulation is responsible for tracking whatever metrics matter.

Key fields:

| Field | Type | Meaning |
|-------|------|---------|
| `Genotype` | `*Genome` | The genetic blueprint |
| `Species` | `*Species` | The species this organism belongs to |
| `Generation` | `int` | Simulation tick at birth (set by the sim on reproduction) |
| `Data` | `*OrganismData` | Application-specific payload (energy, position, age, etc.) |
| `Flag` | `int` | General-purpose integer flag |

The `Data` field is the idiomatic place to attach simulation state:

```go
type CreatureState struct {
    Energy   float64
    Position [2]float64
    Age      int
}

org.Data = &genetics.OrganismData{Value: &CreatureState{Energy: 100.0}}

// retrieve it later:
state := org.Data.Value.(*CreatureState)
state.Energy -= 1.0
```

---

## Genes and Innovation Numbers

Each connection gene carries an **innovation number** — a globally unique integer assigned when that connection first appeared anywhere in the population. Innovation numbers serve as "timestamps" in evolutionary history.

When two organisms mate (sexual reproduction), their genes are aligned by innovation number. Genes present in both parents are **homologous** (matching). Genes present in only one parent are **disjoint** (in the middle) or **excess** (at the end). By default, the primary parent's disjoint and excess genes are inherited.

This alignment mechanism is what makes crossover between genomes of different topologies coherent.

```
Parent 1:  [1][2][3][5][7]
Parent 2:  [1][2][3][4][5][6]
                     ↑ disjoint    ↑ excess from parent 2

Baby:      [1][2][3][5][7]    (parent 1 is primary → excess from parent 1)
```

---

## Species

Species are groups of genetically similar organisms. NEAT uses speciation to **protect innovation**: a new organism that tries a radically different structure would compete poorly against established, well-tuned organisms if it had to compete directly. Placing it in its own species gives it time to improve.

### Compatibility Distance

Two genomes are placed in the same species if their **compatibility distance** is below `CompatThreshold`:

```
distance = DisjointCoeff × (disjoint_genes / max_genes)
         + ExcessCoeff  × (excess_genes  / max_genes)
         + MutdiffCoeff × avg_weight_difference_of_matching_genes
```

Lower threshold → more species, more specialization.
Higher threshold → fewer species, more competition.

In ALife mode, species form automatically whenever a new organism is reproduced. You do not need to manually manage species.

---

## Mutations

Structural mutations change the topology of the network. Weight mutations change the strength of existing connections.

### Structural Mutations (triggered by probability)

| Mutation | Option | Effect |
|----------|--------|--------|
| Add Node | `MutateAddNodeProb` | Splits an existing connection into two, inserting a hidden neuron |
| Add Link | `MutateAddLinkProb` | Adds a new connection between two previously unconnected neurons |
| Connect Sensors | `MutateConnectSensors` | Connects any input nodes that have no outgoing connections |

### Weight/Parameter Mutations (applied when no structural mutation occurs)

| Mutation | Option | Effect |
|----------|--------|--------|
| Link Weights | `MutateLinkWeightsProb` | Gaussian perturbation of all connection weights |
| Toggle Enable | `MutateToggleEnableProb` | Enables or disables a gene |
| Gene Re-enable | `MutateGeneReenableProb` | Re-enables a previously disabled gene |
| Random Trait | `MutateRandomTraitProb` | Perturbs a random trait parameter |
| Link Trait | `MutateLinkTraitProb` | Re-assigns a gene's trait |
| Node Trait | `MutateNodeTraitProb` | Re-assigns a node's trait |

---

## Crossover (Mating)

Sexual reproduction combines two parent genomes. Three crossover strategies are available, selected by probability:

| Strategy | Option | Description |
|----------|--------|-------------|
| Multipoint | `MateMultipointProb` | Each matching gene randomly picked from either parent; non-matching from primary parent |
| Multipoint Average | `MateMultipointAvgProb` | Matching gene weights averaged; non-matching from primary parent |
| Single Point | `MateSinglepointProb` | Genes before a random cut from parent 1, after from parent 2 |

After mating, a further mutation step may be applied unless `MateOnlyProb` prevents it.

---

## The Population

The `Population` is the central data structure. It owns all organisms and all species, and provides thread-safe birth/death operations. It also manages innovation tracking.

In ALife mode, population size is variable — it grows with each reproduction and shrinks with each kill. There is no fixed size constraint enforced by the library.

The `alife.Simulation` wraps the Population and is the primary interface you interact with. Direct access to `Population` is available as `sim.Population` for introspection, but mutations must go through `Simulation` methods.
