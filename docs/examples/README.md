# Examples

These example programs demonstrate different aspects of rt-goNEAT. Each file has a `//go:build ignore` tag so it won't interfere with `go build ./...`. Run them individually with `go run`.

All examples assume they are run from the `docs/examples/` directory and that `../../data/xorstartgenes` is the path to the seed genome.

---

## 01_minimal.go — Minimal Simulation

The smallest possible rt-goNEAT program. Loads a genome, creates a simulation, runs a loop with random reproduction and death, and activates each network.

**Demonstrates:** `NewSimulation`, `ReproduceSexual`, `Kill`, `Phenotype`, `LoadSensors`, `Activate`, `ReadOutputs`

```bash
go run 01_minimal.go
```

---

## 02_energy_lifecycle.go — Energy-Based Lifecycle

A more realistic ALife pattern. Each organism has an energy level, loses energy from metabolism (with a complexity tax for more nodes), gains energy by eating food items, reproduces when energy is high, and dies when it reaches zero.

**Demonstrates:**
- Storing simulation state in `organism.Data`
- Network outputs driving organism movement
- Complexity tax (more hidden nodes → higher metabolic cost)
- Observer callbacks for species events

```bash
go run 02_energy_lifecycle.go
```

---

## 03_serialisation.go — Saving and Restoring

Demonstrates how to checkpoint a running simulation by writing the population to a plain-text file and loading it back. Also shows how to save individual champion genomes.

**Demonstrates:** `Population.Write`, `ReadPopulation`, `species.FindOldest`, `genome.Write`

```bash
go run 03_serialisation.go
```

---

## 04_parallel_activation.go — Parallel Network Activation

Shows the recommended pattern for large populations: activate all networks in parallel across CPU cores, then perform reproduction/death sequentially on the main goroutine.

**Demonstrates:** worker pool with `sync.WaitGroup`, `sim.Organisms()` snapshot, mixing parallel reads with sequential writes

```bash
go run 04_parallel_activation.go
```
