// Package neat implements the NeuroEvolution of Augmenting Topologies (NEAT) method, which can be used to evolve
// specific Artificial Neural Networks from scratch using genetic algorithms.
package neat

import (
	"context"
	"fmt"
	"github.com/Kolterdyx/rt-goNEAT/v4/neat/math"
	"github.com/pkg/errors"
)

var (
	ErrNoActivatorsRegistered                = errors.New("no node activators registered with NEAT options, please assign at least one to NodeActivators")
	ErrActivatorsProbabilitiesNumberMismatch = errors.New("number of node activator probabilities doesn't match number of activators")
)

// GenomeCompatibilityMethod defines the method to calculate genomes compatibility
type GenomeCompatibilityMethod string

const (
	GenomeCompatibilityMethodLinear GenomeCompatibilityMethod = "linear"
	GenomeCompatibilityMethodFast   GenomeCompatibilityMethod = "fast"
)

// Validate checks if this genome compatibility method is supported.
func (g GenomeCompatibilityMethod) Validate() error {
	if g != GenomeCompatibilityMethodLinear && g != GenomeCompatibilityMethodFast {
		return errors.Errorf("unsupported genome compatibility method: [%s]", g)
	}
	return nil
}

// Options holds the NEAT algorithm parameters.
type Options struct {
	// Probability of mutating a single trait param
	TraitParamMutProb float64 `yaml:"trait_param_mut_prob"`
	// Power of mutation on a single trait param
	TraitMutationPower float64 `yaml:"trait_mutation_power"`
	// The power of a link weight mutation
	WeightMutPower float64 `yaml:"weight_mut_power"`

	// Genome compatibility coefficients.
	// Compatibility = disjoint_coeff * pdg + excess_coeff * peg + mutdiff_coeff * mdmg
	DisjointCoeff float64 `yaml:"disjoint_coeff"`
	ExcessCoeff   float64 `yaml:"excess_coeff"`
	MutdiffCoeff  float64 `yaml:"mutdiff_coeff"`

	// CompatThreshold is the compatibility distance below which two genomes are considered the same species.
	CompatThreshold float64 `yaml:"compat_threshold"`

	// Probabilities of a non-mating reproduction
	MutateOnlyProb         float64 `yaml:"mutate_only_prob"`
	MutateRandomTraitProb  float64 `yaml:"mutate_random_trait_prob"`
	MutateLinkTraitProb    float64 `yaml:"mutate_link_trait_prob"`
	MutateNodeTraitProb    float64 `yaml:"mutate_node_trait_prob"`
	MutateLinkWeightsProb  float64 `yaml:"mutate_link_weights_prob"`
	MutateToggleEnableProb float64 `yaml:"mutate_toggle_enable_prob"`
	MutateGeneReenableProb float64 `yaml:"mutate_gene_reenable_prob"`
	MutateAddNodeProb      float64 `yaml:"mutate_add_node_prob"`
	MutateAddLinkProb      float64 `yaml:"mutate_add_link_prob"`
	// Probability of mutation involving disconnected input connections
	MutateConnectSensors float64 `yaml:"mutate_connect_sensors"`

	// Probabilities for cross-species mating and crossover type selection
	InterspeciesMateRate  float64 `yaml:"interspecies_mate_rate"`
	MateMultipointProb    float64 `yaml:"mate_multipoint_prob"`
	MateMultipointAvgProb float64 `yaml:"mate_multipoint_avg_prob"`
	MateSinglepointProb   float64 `yaml:"mate_singlepoint_prob"`

	// MateOnlyProb is the probability of mating without subsequent mutation
	MateOnlyProb float64 `yaml:"mate_only_prob"`
	// RecurOnlyProb forces selection of only recurrent links when adding a link
	RecurOnlyProb float64 `yaml:"recur_only_prob"`

	// PopSize is the initial population size (population is variable-size in ALife mode)
	PopSize int `yaml:"pop_size"`
	// NewLinkTries is the number of attempts mutateAddLink makes to find an unconnected pair
	NewLinkTries int `yaml:"newlink_tries"`

	// GenCompatMethod selects the genome compatibility calculation (linear or fast)
	GenCompatMethod GenomeCompatibilityMethod `yaml:"genome_compat_method"`

	// NodeActivators is the list of activation functions to choose from for new nodes
	NodeActivators []math.NodeActivationType `yaml:"-"`
	// NodeActivatorsProb are the probabilities of each activator in NodeActivators
	NodeActivatorsProb []float64 `yaml:"-"`

	// NodeActivatorsWithProbs is the YAML representation of NodeActivators+Probs
	NodeActivatorsWithProbs []string `yaml:"node_activators"`

	// LogLevel controls log output verbosity
	LogLevel string `yaml:"log_level"`
}

// RandomNodeActivationType returns a random activation type from the registered set.
func (c *Options) RandomNodeActivationType() (math.NodeActivationType, error) {
	if len(c.NodeActivators) == 0 {
		return 0, ErrNoActivatorsRegistered
	}
	if len(c.NodeActivators) == 1 {
		return c.NodeActivators[0], nil
	}
	if len(c.NodeActivators) != len(c.NodeActivatorsProb) {
		return 0, ErrActivatorsProbabilitiesNumberMismatch
	}
	index := math.SingleRouletteThrow(c.NodeActivatorsProb)
	if index < 0 || index >= len(c.NodeActivators) {
		return 0, fmt.Errorf("unexpected error when trying to find random node activator, activator index: %d", index)
	}
	return c.NodeActivators[index], nil
}

// Validate checks that the options are internally consistent.
func (c *Options) Validate() error {
	if err := c.GenCompatMethod.Validate(); err != nil {
		return err
	}
	if len(c.NodeActivators) == 0 {
		return ErrNoActivatorsRegistered
	}
	if len(c.NodeActivators) != len(c.NodeActivatorsProb) {
		return ErrActivatorsProbabilitiesNumberMismatch
	}
	return nil
}

// NeatContext returns a context carrying these options.
func (c *Options) NeatContext() context.Context {
	return NewContext(context.Background(), c)
}
