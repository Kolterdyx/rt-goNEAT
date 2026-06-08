# Configuration

All NEAT algorithm parameters are controlled through `neat.Options`. You can build options in code or load them from a file.

---

## Loading from a file

### Auto-detection

```go
opts, err := neat.ReadNeatOptionsFromFile("config/world.neat.yml") // YAML
opts, err := neat.ReadNeatOptionsFromFile("config/world.neat")     // plain text
```

The file format is detected from the extension: `.yml` / `.yaml` → YAML, anything else → plain text.

### Explicit loaders

```go
f, _ := os.Open("config/world.neat.yml")
opts, err := neat.LoadYAMLOptions(f)

f, _ := os.Open("config/world.neat")
opts, err := neat.LoadNeatOptions(f)
```

---

## Plain-text format (.neat)

One `key value` pair per line. Unknown keys and legacy epoch parameters (`num_runs`, `num_generations`, `epoch_executor`, `dropoff_age`, `babies_stolen`, `print_every`, `age_significance`, `survival_thresh`) are silently ignored so that existing config files continue to load.

```
trait_param_mut_prob   0.5
trait_mutation_power   1.0
weight_mut_power       2.5
disjoint_coeff         1.0
excess_coeff           1.0
mutdiff_coeff          0.4
compat_threshold       3.0
mutate_only_prob       0.25
mutate_random_trait_prob 0.1
mutate_link_trait_prob 0.1
mutate_node_trait_prob 0.1
mutate_link_weights_prob 0.9
mutate_toggle_enable_prob 0.0
mutate_gene_reenable_prob 0.0
mutate_add_node_prob   0.03
mutate_add_link_prob   0.08
mutate_connect_sensors 0.5
interspecies_mate_rate 0.001
mate_multipoint_prob   0.3
mate_multipoint_avg_prob 0.3
mate_singlepoint_prob  0.3
mate_only_prob         0.2
recur_only_prob        0.0
pop_size               100
newlink_tries          50
genome_compat_method   linear
log_level              info
```

---

## YAML format (.neat.yml)

```yaml
trait_param_mut_prob: 0.5
trait_mutation_power: 1.0
weight_mut_power: 2.5

disjoint_coeff: 1.0
excess_coeff: 1.0
mutdiff_coeff: 0.4
compat_threshold: 3.0

mutate_only_prob: 0.25
mutate_random_trait_prob: 0.1
mutate_link_trait_prob: 0.1
mutate_node_trait_prob: 0.1
mutate_link_weights_prob: 0.9
mutate_toggle_enable_prob: 0.0
mutate_gene_reenable_prob: 0.0
mutate_add_node_prob: 0.03
mutate_add_link_prob: 0.08
mutate_connect_sensors: 0.5

interspecies_mate_rate: 0.001
mate_multipoint_prob: 0.3
mate_multipoint_avg_prob: 0.3
mate_singlepoint_prob: 0.3
mate_only_prob: 0.2
recur_only_prob: 0.0

pop_size: 100
newlink_tries: 50
genome_compat_method: fast
log_level: info

# Activation functions for new hidden nodes:
# each entry is "FunctionName Probability"
node_activators:
  - "SigmoidSteepenedActivation 0.25"
  - "TanhActivation 0.25"
  - "GaussianBipolarActivation 0.25"
  - "LinearAbsActivation 0.25"
```

---

## All Parameters

### Genome compatibility (speciation)

| YAML key | Go field | Description |
|----------|----------|-------------|
| `disjoint_coeff` | `DisjointCoeff` | Weight of disjoint gene count in compatibility formula |
| `excess_coeff` | `ExcessCoeff` | Weight of excess gene count |
| `mutdiff_coeff` | `MutdiffCoeff` | Weight of average weight difference in matching genes |
| `compat_threshold` | `CompatThreshold` | Maximum compatibility distance for same-species assignment |
| `genome_compat_method` | `GenCompatMethod` | `"linear"` (O(n), exact) or `"fast"` (approximate, better for large genomes) |

**Tuning tip:** Start with `compat_threshold: 3.0`. Lower it (1.5–2.5) to get more species and stronger niche protection. Raise it (4–6) for a single melting-pot population.

---

### Mutation probabilities

These probabilities apply to each newly created organism. They are checked in order; the first structural mutation that fires short-circuits the rest. If no structural mutation fires, all non-structural mutations are applied.

| YAML key | Go field | Effect |
|----------|----------|--------|
| `mutate_add_node_prob` | `MutateAddNodeProb` | Probability of splitting a connection and inserting a hidden neuron |
| `mutate_add_link_prob` | `MutateAddLinkProb` | Probability of adding a new connection |
| `mutate_connect_sensors` | `MutateConnectSensors` | Probability of connecting a disconnected input node |
| `mutate_link_weights_prob` | `MutateLinkWeightsProb` | Probability of perturbing all connection weights |
| `mutate_toggle_enable_prob` | `MutateToggleEnableProb` | Probability of toggling gene enable/disable |
| `mutate_gene_reenable_prob` | `MutateGeneReenableProb` | Probability of re-enabling a disabled gene |
| `mutate_random_trait_prob` | `MutateRandomTraitProb` | Probability of mutating a random trait parameter |
| `mutate_link_trait_prob` | `MutateLinkTraitProb` | Probability of reassigning a link's trait |
| `mutate_node_trait_prob` | `MutateNodeTraitProb` | Probability of reassigning a node's trait |
| `mutate_only_prob` | `MutateOnlyProb` | Probability that asexual reproduction is used (vs. mating) |
| `weight_mut_power` | `WeightMutPower` | Standard deviation of Gaussian weight perturbation |

**Tuning tip:** For early-stage evolution keep `mutate_add_node_prob` and `mutate_add_link_prob` small (0.01–0.05). Weight mutations (`mutate_link_weights_prob: 0.8`) should dominate early on.

---

### Mating (crossover) parameters

| YAML key | Go field | Description |
|----------|----------|-------------|
| `mate_multipoint_prob` | `MateMultipointProb` | Probability of multipoint crossover (randomly picks from each parent at each gene) |
| `mate_multipoint_avg_prob` | `MateMultipointAvgProb` | Probability of averaged multipoint crossover |
| `mate_singlepoint_prob` | `MateSinglepointProb` | Probability of single-point crossover |
| `mate_only_prob` | `MateOnlyProb` | Probability that the offspring is NOT mutated after mating |
| `interspecies_mate_rate` | `InterspeciesMateRate` | Probability that the second parent is from a different species |
| `recur_only_prob` | `RecurOnlyProb` | When adding a link, probability that it must be recurrent |

The three `mate_*_prob` values do not need to sum to 1. They are evaluated as: if `rand < MateMultipointProb` → multipoint; else if `rand < MateMultipointAvgProb / (MateMultipointAvgProb + MateSinglepointProb)` → avg; else → singlepoint.

---

### Population and search parameters

| YAML key | Go field | Description |
|----------|----------|-------------|
| `pop_size` | `PopSize` | Number of organisms in the initial population (population is variable-size after seeding) |
| `newlink_tries` | `NewLinkTries` | Number of random node-pair attempts when adding a new link |

---

### Activation functions

New hidden neurons added by `mutate_add_node` are assigned a random activation function from the `NodeActivators` list.

**In code:**

```go
import neatmath "github.com/Kolterdyx/rt-goNEAT/v4/neat/math"

opts.NodeActivators = []neatmath.NodeActivationType{
    neatmath.SigmoidSteepenedActivation,
    neatmath.TanhActivation,
    neatmath.GaussianBipolarActivation,
}
opts.NodeActivatorsProb = []float64{0.5, 0.3, 0.2}
```

**In YAML:**

```yaml
node_activators:
  - "SigmoidSteepenedActivation 0.5"
  - "TanhActivation 0.3"
  - "GaussianBipolarActivation 0.2"
```

### Available activation functions

| Name | Description |
|------|-------------|
| `SigmoidSteepenedActivation` | Standard sigmoid with steepened slope (default) |
| `SigmoidPlainActivation` | Standard logistic sigmoid 1/(1+e^-x) |
| `SigmoidBipolarActivation` | Bipolar sigmoid: output in [-1, 1] |
| `TanhActivation` | Hyperbolic tangent |
| `GaussianActivation` | Gaussian bell curve |
| `GaussianBipolarActivation` | Bipolar Gaussian |
| `LinearActivation` | f(x) = x |
| `LinearAbsActivation` | f(x) = |x| |
| `LinearClippedActivation` | f(x) = clamp(x, 0, 1) |
| `SineActivation` | f(x) = sin(x) |
| `StepActivation` | f(x) = 0 or 1 |
| `SignActivation` | f(x) = -1, 0, or 1 |
| `NullActivation` | f(x) = 0 (useful for bias nodes) |

---

### Log levels

| `log_level` | Effect |
|-------------|--------|
| `"debug"` | Verbose: every speciation decision, every mutation attempt |
| `"info"` | Normal operation messages |
| `"warn"` | Warnings only (e.g., unusual genome states) |
| `"error"` | Errors only |
| `""` (empty) | Disables logging |

---

## Validating options

```go
if err := opts.Validate(); err != nil {
    log.Fatal("invalid NEAT options:", err)
}
```

`Validate` checks that:
- `GenCompatMethod` is `"linear"` or `"fast"`
- At least one `NodeActivator` is registered
- `NodeActivators` and `NodeActivatorsProb` have the same length

Note: `Validate` is called automatically by `LoadNeatOptions` and `LoadYAMLOptions`.
