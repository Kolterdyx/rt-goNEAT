# Neural Networks

Each organism's genome describes a neural network topology. When you need the actual network to process inputs, call `org.Phenotype()` to get the `*network.Network`. The network is built lazily from the genome on first access and cached for the organism's lifetime.

```go
import "github.com/Kolterdyx/rt-goNEAT/v4/neat/network"
```

---

## Getting the Network

```go
net, err := org.Phenotype()
if err != nil {
    // genesis failed — malformed genome
}
```

The network is rebuilt only when you explicitly call `org.UpdatePhenotype()`, which you should do after manually modifying the genome. For organisms created through the normal reproduction path, you never need to call this.

---

## Activating the Network

A complete activation cycle is:

```go
// 1. Load input values
err := net.LoadSensors([]float64{0.5, 1.0, -0.3})

// 2. Propagate activation
activated, err := net.Activate()
if !activated {
    // network did not converge (recurrent networks may need more steps)
}

// 3. Read output values
outputs := net.ReadOutputs()
```

### How activation works

The network uses **iterative activation**. On each step, every non-input node computes:

```
activationSum = Σ (incomingWeight × sourceNodeActivation)
nodeOutput    = activationFunction(activationSum)
```

This repeats until all output nodes have been activated. For feedforward networks, this typically converges in one or two passes. Recurrent networks may require more steps.

`Activate()` calls `ActivateSteps(20)` internally. For complex recurrent networks or unusual topologies, you can call `ActivateSteps(n)` with a larger `n`:

```go
activated, err := net.ActivateSteps(50)
```

For feedforward networks that are guaranteed to have no cycles, there is also a faster solver:

```go
solver, err := net.FastNetworkSolver()
if err != nil {
    // fallback to standard activation
    net.Activate()
} else {
    activated, err := solver.ForwardSteps(1)
}
```

---

## Input and Output Sizes

The number of inputs and outputs is fixed by the seed genome. You can query them at runtime:

```go
fmt.Printf("nodes: %d, links: %d\n", net.NodeCount(), net.LinkCount())
```

The input count is implicit in the genome's sensor nodes. To know the number of inputs and outputs your seed genome expects, count the sensor/output nodes in the genome file or inspect them:

```go
for _, node := range genome.Nodes {
    switch node.NeuronType {
    case network.InputNeuron:
        // input slot
    case network.OutputNeuron:
        // output slot
    case network.BiasNeuron:
        // always outputs 1.0; LoadSensors does not touch bias nodes
    case network.HiddenNeuron:
        // evolved hidden node
    }
}
```

`LoadSensors` loads values into sensor (input) nodes in order, ignoring bias nodes. If you pass fewer values than sensor nodes exist, only the first `len(inputs)` sensors are loaded — the rest retain their previous value.

---

## Resetting Network State

For recurrent networks it may be desirable to reset all activation values between independent episodes:

```go
ok, err := net.Flush()
```

This zeroes all node activations. After flushing, the next `Activate()` call starts from scratch.

---

## The Solver Interface

`network.Network` implements the `network.Solver` interface, which defines the full set of network operations:

```go
type Solver interface {
    ForwardSteps(steps int) (bool, error)   // iterative forward propagation
    RecursiveSteps() (bool, error)           // recursive from outputs
    Relax(maxSteps int, delta float64) (bool, error)
    Flush() (bool, error)
    LoadSensors(inputs []float64) error
    ReadOutputs() []float64
    NodeCount() int
    LinkCount() int
}
```

You can use the `Solver` interface type in your simulation code so that the same logic works with both the standard `*Network` and a `FastNetworkSolver`:

```go
func evaluateOrganism(solver network.Solver, inputs []float64) []float64 {
    solver.LoadSensors(inputs)
    solver.ForwardSteps(1)
    return solver.ReadOutputs()
}
```

---

## Efficient Batch Activation

For large populations, activate networks in parallel across goroutines. Use `sim.Organisms()` to get a safe snapshot first, then spawn workers:

```go
orgs := sim.Organisms()
results := make([][]float64, len(orgs))

var wg sync.WaitGroup
for i, org := range orgs {
    wg.Add(1)
    go func(idx int, o *genetics.Organism) {
        defer wg.Done()
        net, err := o.Phenotype()
        if err != nil {
            return
        }
        _ = net.LoadSensors(getInputsFor(o))
        net.Activate()
        results[idx] = net.ReadOutputs()
    }(i, org)
}
wg.Wait()

// Now use results[i] alongside orgs[i]
```

Each organism's phenotype is independent — there are no shared data structures between different organisms' networks, so goroutine-level parallelism is safe.

---

## Accessing Network Internals

For advanced inspection:

```go
net, _ := org.Phenotype()

// Output nodes
for _, n := range net.Outputs {
    fmt.Printf("output node %d: activation=%.4f\n", n.Id, n.Activation)
}

// All nodes (via the genome)
for _, node := range org.Genotype.Nodes {
    fmt.Printf("node %d type=%v activation=%v\n",
        node.Id, node.NeuronType, node.ActivationType)
}

// All links (genes)
for _, gene := range org.Genotype.Genes {
    if gene.IsEnabled {
        fmt.Printf("  %d→%d weight=%.4f\n",
            gene.Link.InNode.Id, gene.Link.OutNode.Id, gene.Link.ConnectionWeight)
    }
}
```

---

## Network Topology Over Time

Organisms mutate their network topology as they reproduce. An organism born with the minimal 3-input → 1-output network may, after several generations of mutation, have multiple hidden layers.

To track topology growth:

```go
for _, org := range sim.Organisms() {
    genes := org.Genotype.Extrons() // number of enabled genes
    nodes := len(org.Genotype.Nodes)
    fmt.Printf("org species=%d nodes=%d links=%d\n", org.Species.Id, nodes, genes)
}
```

`Extrons()` returns the count of enabled genes (active connections). Disabled genes are retained in the genome for crossover alignment purposes but do not affect network computation.
