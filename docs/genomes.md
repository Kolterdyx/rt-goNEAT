# Genomes

A `Genome` encodes the complete blueprint of an organism's neural network. It contains a list of nodes (neurons), a list of genes (connections), and a list of traits (evolvable parameters).

```go
import "github.com/Kolterdyx/rt-goNEAT/v4/neat/genetics"
```

---

## Genome Structure

```go
type Genome struct {
    Id           int
    Traits       []*neat.Trait           // evolvable parameter groups
    Nodes        []*network.NNode        // all neurons (input, hidden, output, bias)
    Genes        []*Gene                 // connections with innovation numbers
    ControlGenes []*MIMOControlGene      // MIMO module genes (advanced)
    Phenotype    *network.Network        // cached phenotype (nil until Genesis called)
}
```

### Nodes

Nodes are typed:

| Type | Constant | Meaning |
|------|----------|---------|
| Input / Sensor | `network.InputNeuron` | Receives external input values |
| Output | `network.OutputNeuron` | Produces the network's output |
| Bias | `network.BiasNeuron` | Always outputs 1.0; provides a constant offset |
| Hidden | `network.HiddenNeuron` | Internal computation nodes (evolved via mutation) |

The minimal genome has only input, output, and bias nodes with direct connections from inputs to outputs.

### Genes

Each gene represents one directed connection:

```go
type Gene struct {
    Link         *network.Link   // connection (InNode → OutNode, weight)
    InnovationNum int64          // evolutionary timestamp
    MutationNum  float64         // cumulative weight change history
    IsEnabled    bool            // disabled genes are kept for crossover alignment
}
```

Disabled genes retain their history and can be re-enabled by the `mutateGeneReEnable` mutation.

### Traits

Traits are groups of 8 evolvable floating-point parameters associated with nodes or genes. They are mutated by the trait mutation operators. In basic usage you can ignore them; they become relevant when writing custom activation functions that read `node.Params`.

---

## Plain-Text Genome Format

### Example genome file

```
genomestart 1
trait 1 0.1 0 0 0 0 0 0 0
trait 2 0.2 0 0 0 0 0 0 0
trait 3 0.3 0 0 0 0 0 0 0
node 1 0 1 1
node 2 0 1 1
node 3 0 1 3
node 4 0 0 2
gene 1 1 4 1.5 false 1 0 true
gene 2 2 4 2.5 false 2 0 true
gene 3 3 4 3.5 false 3 0 true
genomeend 1
```

### Syntax

**`genomestart <id>`** — begins a genome block; `<id>` is the genome integer ID.

**`trait <id> <p1> <p2> ... <p8>`** — a trait with 8 float parameters.

**`node <id> <traitId> <type> <neuronType> [<activationFn>]`**

| Column | Values |
|--------|--------|
| `id` | Unique node integer ID |
| `traitId` | Trait to associate (0 = no trait) |
| `type` | `0` = neuron, `1` = sensor |
| `neuronType` | `1` = bias, `2` = output, `3` = input/hidden, `4` = hidden |
| `activationFn` | Optional: activation function name (e.g. `SigmoidSteepenedActivation`) |

**`gene <traitId> <inNodeId> <outNodeId> <weight> <isRecurrent> <innovNum> <mutNum> <isEnabled>`**

| Column | Values |
|--------|--------|
| `traitId` | Trait ID for this gene (0 = first trait) |
| `inNodeId` | Source node ID |
| `outNodeId` | Target node ID |
| `weight` | Connection weight (float) |
| `isRecurrent` | `true` or `false` |
| `innovNum` | Innovation number (globally unique) |
| `mutNum` | Mutation history number |
| `isEnabled` | `true` or `false` |

**`genomeend <id>`** — closes the genome block.

---

## YAML Genome Format

The YAML format is richer and recommended for hand-crafted genomes:

```yaml
id: 1
traits:
  - id: 1
    params: [0.1, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0]
nodes:
  - id: 1
    trait_id: 1
    neuron_type: 1   # bias
    activation: SigmoidSteepenedActivation
  - id: 2
    trait_id: 1
    neuron_type: 3   # input
    activation: NullActivation
  - id: 3
    trait_id: 1
    neuron_type: 3   # input
    activation: NullActivation
  - id: 4
    trait_id: 1
    neuron_type: 2   # output
    activation: SigmoidSteepenedActivation
genes:
  - trait_id: 1
    src_id: 1
    dst_id: 4
    weight: 1.0
    recurrent: false
    innov_num: 1
    mut_num: 0.0
    enabled: true
  - trait_id: 1
    src_id: 2
    dst_id: 4
    weight: 0.5
    recurrent: false
    innov_num: 2
    mut_num: 0.0
    enabled: true
  - trait_id: 1
    src_id: 3
    dst_id: 4
    weight: -0.5
    recurrent: false
    innov_num: 3
    mut_num: 0.0
    enabled: true
```

---

## Reading a Genome

### From a file (auto-detect format)

```go
reader, err := genetics.NewGenomeReaderFromFile("data/mystartgenes")
if err != nil {
    log.Fatal(err)
}
genome, err := reader.Read()
```

### From a reader

```go
// Plain text
reader := genetics.NewGenomeReader(strings.NewReader(plainTextGenome), genetics.PlainGenomeEncoding)
genome, err := reader.Read()

// YAML
reader = genetics.NewGenomeReader(strings.NewReader(yamlGenome), genetics.YAMLGenomeEncoding)
genome, err = reader.Read()
```

### Directly from a raw reader (population file)

```go
f, _ := os.Open("population.neat")
pop, err := genetics.ReadPopulation(f, opts)
```

A population file is multiple genomes concatenated.

---

## Writing a Genome

```go
// To a file
f, _ := os.Create("saved_genome.neat")
err := genome.Write(f)

// To a string
var buf bytes.Buffer
err := genome.Write(&buf)
```

### Writing the whole population

```go
// All genomes sequentially
err := pop.Write(f)

// Grouped by species with headers
err := pop.WriteBySpecies(f)
```

---

## Designing Your Seed Genome

### Rules

1. **Every output node must be reachable** from at least one input via a direct gene. The minimal genome has `n_inputs + 1 bias → n_outputs` connections.

2. **Innovation numbers must be globally unique and increasing** across all genes in the file. They don't need to be sequential; gaps are fine.

3. **Node IDs must be unique** within the genome.

4. **Start minimal.** The NEAT algorithm grows complexity as needed. Starting with a deeply connected genome wastes capacity and slows evolution.

### Minimal feedforward template (2 inputs → 1 output)

```
genomestart 1
trait 1 0.0 0 0 0 0 0 0 0
node 1 1 1 1       <- bias node (type=1, neuronType=1)
node 2 1 1 3       <- input node (type=1, neuronType=3)
node 3 1 1 3       <- input node (type=1, neuronType=3)
node 4 1 0 2       <- output node (type=0, neuronType=2)
gene 1 1 4 0.0 false 1 0 true    <- bias→output
gene 1 2 4 0.0 false 2 0 true    <- input1→output
gene 1 3 4 0.0 false 3 0 true    <- input2→output
genomeend 1
```

Starting with weights of 0.0 lets the first generation of mutations set the initial values.

### Disconnected sensors

If you want NEAT to discover which inputs are relevant, use `xordisconnectedstartgenes` as a template — input nodes are present in the genome but have no outgoing genes. The `mutate_connect_sensors` option will gradually connect them.

---

## Building a Genome Programmatically

```go
import (
    "github.com/Kolterdyx/rt-goNEAT/v4/neat"
    "github.com/Kolterdyx/rt-goNEAT/v4/neat/genetics"
    "github.com/Kolterdyx/rt-goNEAT/v4/neat/network"
    neatmath "github.com/Kolterdyx/rt-goNEAT/v4/neat/math"
)

trait := &neat.Trait{Id: 1, Params: make([]float64, 8)}

bias   := network.NewSensorNode(1, true)  // bias node
input1 := network.NewSensorNode(2, false)
input2 := network.NewSensorNode(3, false)
output := network.NewNNode(4, network.OutputNeuron)
output.ActivationType = neatmath.SigmoidSteepenedActivation

nodes := []*network.NNode{bias, input1, input2, output}

genes := []*genetics.Gene{
    genetics.NewGeneWithTrait(trait, 0.0, bias,   output, false, 1, 0),
    genetics.NewGeneWithTrait(trait, 0.0, input1, output, false, 2, 0),
    genetics.NewGeneWithTrait(trait, 0.0, input2, output, false, 3, 0),
}

genome := genetics.NewGenome(1, []*neat.Trait{trait}, nodes, genes)
```
