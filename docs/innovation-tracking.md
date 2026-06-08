# Innovation Tracking

Innovation tracking is the mechanism that allows NEAT to perform meaningful crossover between genomes of different topologies. Without it, two genomes that independently evolved the same structural change would assign different IDs to it, causing misalignment during mating.

---

## Why Innovations Matter

When two parent genomes mate, their genes are aligned by **innovation number**. Genes with the same innovation number are "homologous" — they represent the same structural event in evolutionary history.

Consider two parent genomes:

```
Parent A:  gene1(innov=1) gene2(innov=2) gene3(innov=3) gene5(innov=5)
Parent B:  gene1(innov=1) gene2(innov=2) gene3(innov=3) gene4(innov=4) gene5(innov=5)
```

Gene 4 is present only in Parent B. Because of innovation numbers, the library knows exactly where it belongs. If Parent A is the primary parent (dominant), gene 4 is dropped; if Parent B is primary, gene 4 is included.

Without innovation numbers, combining two topologically different genomes would be a random shuffle with no guaranteed structure.

---

## The InnovationTracker

rt-goNEAT uses a map-based `InnovationTracker` for O(1) innovation lookup. This is critical for long-running ALife simulations where the innovation history would otherwise grow unbounded and make each structural mutation O(n) slower.

Each population owns one `InnovationTracker`:

```
Population.tracker (*InnovationTracker)
    ├── byKey: map[innovationKey]*Innovation    — O(1) lookup
    ├── nextInnovNum: int64                     — atomic counter
    └── nextNodeId:   int32                     — atomic counter
```

The tracker implements both `InnovationsObserver` (for genome mutation operators) and `network.NodeIdGenerator` (for new hidden node IDs).

---

## How Innovations Are Assigned

When `mutateAddNode` or `mutateAddLink` fires:

1. The mutation operator calls `innovations.FindLinkInnovation(...)` or `innovations.FindNodeInnovation(...)` — an O(1) map lookup.
2. If an identical structural change was already made by another organism during this simulation window: the **same** innovation number is reused. This ensures the baby's gene aligns correctly with its sibling at mating time.
3. If this is a new structural change: `innovations.NextInnovationNumber()` generates a new ID, and the innovation is stored.

The tracker is keyed by:

| Change type | Key components |
|-------------|----------------|
| New link | `(inNodeId, outNodeId, isRecurrent)` |
| New node (split) | `(inNodeId, outNodeId, oldInnovNum)` |

---

## Persistent Innovations

Unlike epoch-based NEAT where innovations are cleared each generation, rt-goNEAT's `InnovationTracker` persists forever (or until you explicitly clear it). This is correct for ALife simulations — there is no generation boundary to mark the start of a new innovation window.

**Implication:** The same structural change (e.g., a link from node 3 to node 7) will always get the same innovation number for the lifetime of the simulation, regardless of how many ticks have passed. This makes long-term lineage tracking coherent.

---

## When to Call ClearInnovations

Clearing innovations should be rare. The only situation where it makes sense is:

- You intentionally want to allow the same structural topology to evolve independently in different lineages and treat them as *different* innovations going forward
- You are implementing a "speciation reset" event in your simulation

```go
sim.Population.ClearInnovations()
```

This clears the innovation map but does **not** reset the innovation number counter — new innovations will continue numbering from where they left off. Existing gene innovation numbers in living organisms are unaffected.

---

## Accessing Innovation Data

```go
// All recorded innovations
innovations := sim.Population.Innovations() // []Innovation

for _, inn := range innovations {
    switch {
    case inn.InNodeId == 0:
        // node-split innovation
        fmt.Printf("node-split: %d→%d → new node %d\n",
            inn.InNodeId, inn.OutNodeId, inn.NewNodeId)
    default:
        // link innovation
        fmt.Printf("new link: %d→%d weight=%.3f recurrent=%v innov=%d\n",
            inn.InNodeId, inn.OutNodeId, inn.NewWeight, inn.IsRecurrent, inn.InnovationNum)
    }
}

// Current counters
nextInnov := sim.Population.NextInnovationNumber() // peek (and increment)
nextNode  := sim.Population.NextNodeId()           // peek (and increment)
```

Note: `NextInnovationNumber()` and `NextNodeId()` are **consuming** operations — each call increments the counter. Do not call them for inspection; use them only when you actually need to allocate a new ID.

---

## Tracking Structural Complexity

You can derive structural statistics without inspecting innovations directly:

```go
for _, org := range sim.Organisms() {
    genome := org.Genotype
    enabled := genome.Extrons()           // active connections
    total   := len(genome.Genes)          // total connections (including disabled)
    nodes   := len(genome.Nodes)          // all nodes
    hidden  := 0
    for _, n := range genome.Nodes {
        if n.NeuronType == network.HiddenNeuron {
            hidden++
        }
    }
    fmt.Printf("species=%d nodes=%d hidden=%d links=%d/%d\n",
        org.Species.Id, nodes, hidden, enabled, total)
}
```
