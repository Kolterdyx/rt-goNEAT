# rt-goNEAT Documentation

rt-goNEAT is a real-time Artificial Life simulation library built on the NEAT (NeuroEvolution of Augmenting Topologies) algorithm. Unlike traditional NEAT implementations that evolve populations in synchronized generations, rt-goNEAT gives the simulation complete control over when organisms are born, reproduce, and die.

## Table of Contents

1. [Core Concepts](core-concepts.md) — What NEAT is, how genotypes and phenotypes relate, terminology
2. [Getting Started](getting-started.md) — Installation, your first simulation in 30 lines
3. [Configuration](configuration.md) — All `Options` parameters, YAML and plain-text config files
4. [Simulation API](simulation-api.md) — The `alife.Simulation` controller in depth
5. [Neural Networks](neural-networks.md) — Working with the network phenotype: loading inputs, activating, reading outputs
6. [Genomes](genomes.md) — Genome structure, file formats, loading and saving
7. [Speciation](speciation.md) — How species form, what compatibility distance means, tuning thresholds
8. [Innovation Tracking](innovation-tracking.md) — Why innovations matter, the O(1) tracker, when to call ClearInnovations
9. [Observer Pattern](observer.md) — Reacting to lifecycle events without polling
10. [Examples](examples/) — Complete working programs

## Quick Reference

```
github.com/Kolterdyx/rt-goNEAT/v4/alife        — simulation controller
github.com/Kolterdyx/rt-goNEAT/v4/neat          — options, context
github.com/Kolterdyx/rt-goNEAT/v4/neat/genetics — genome, organism, species, population
github.com/Kolterdyx/rt-goNEAT/v4/neat/network  — neural network phenotype
github.com/Kolterdyx/rt-goNEAT/v4/neat/math     — activation functions
```

## What changed from the original goNEAT

| Old (epoch-based)                      | New (real-time ALife)                          |
|----------------------------------------|------------------------------------------------|
| `experiment.Experiment.Execute()`      | Your own simulation loop                       |
| `experiment.GenerationEvaluator`       | Your own reproduction logic                    |
| `PopulationEpochExecutor.NextEpoch()`  | `sim.ReproduceAsexual()` / `ReproduceSexual()` |
| `organism.Fitness` / `IsWinner`        | Simulation-specific data via `organism.Data`   |
| Fixed population size per generation   | Variable population, grow/shrink freely        |
| Innovations cleared each generation    | Innovations persist; clear manually if needed  |
