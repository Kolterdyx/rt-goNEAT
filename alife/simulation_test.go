package alife

import (
	"context"
	"testing"

	"github.com/Kolterdyx/rt-goNEAT/v4/neat"
	"github.com/Kolterdyx/rt-goNEAT/v4/neat/genetics"
	neatmath "github.com/Kolterdyx/rt-goNEAT/v4/neat/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOpts returns minimal Options sufficient to seed and mutate a population.
func testOpts() *neat.Options {
	return &neat.Options{
		CompatThreshold:       3.0,
		PopSize:               5,
		MutateOnlyProb:        0.25,
		MutateAddNodeProb:     0.01,
		MutateAddLinkProb:     0.3,
		MutateLinkWeightsProb: 0.8,
		MateMultipointProb:    0.6,
		MateMultipointAvgProb: 0.4,
		MateSinglepointProb:   0.0,
		MateOnlyProb:          0.2,
		DisjointCoeff:         1.0,
		ExcessCoeff:           1.0,
		MutdiffCoeff:          0.4,
		WeightMutPower:        2.5,
		NewLinkTries:          20,
		GenCompatMethod:       neat.GenomeCompatibilityMethodLinear,
		NodeActivators:        []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
		NodeActivatorsProb:    []float64{1.0},
	}
}

// loadTestGenome reads the XOR start genome used by many existing tests.
func loadTestGenome(t *testing.T) *genetics.Genome {
	t.Helper()
	reader, err := genetics.NewGenomeReaderFromFile("../data/xorstartgenes")
	require.NoError(t, err, "failed to open test genome file")
	g, err := reader.Read()
	require.NoError(t, err, "failed to read test genome")
	return g
}

// TestNewSimulation verifies that NewSimulation seeds a population of the expected size.
func TestNewSimulation(t *testing.T) {
	g := loadTestGenome(t)
	opts := testOpts()

	sim, err := NewSimulation(context.Background(), g, opts)
	require.NoError(t, err)
	require.NotNil(t, sim)

	orgs := sim.Organisms()
	assert.Len(t, orgs, opts.PopSize)
	assert.Equal(t, int64(0), sim.Tick())
}

// TestSimulation_Step verifies the tick counter increments.
func TestSimulation_Step(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	sim.Step()
	sim.Step()
	assert.Equal(t, int64(2), sim.Tick())
}

// TestSimulation_ReproduceAsexual verifies that asexual reproduction adds an organism.
func TestSimulation_ReproduceAsexual(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	before := len(sim.Organisms())
	parent := sim.Organisms()[0]

	offspring, err := sim.ReproduceAsexual(parent)
	require.NoError(t, err)
	require.NotNil(t, offspring)

	assert.Len(t, sim.Organisms(), before+1, "population should grow by 1 after asexual reproduction")
	assert.True(t, offspring.IsAlive())
	assert.NotNil(t, offspring.Species, "offspring must be speciated")
}

// TestSimulation_ReproduceSexual verifies that sexual reproduction adds an organism.
func TestSimulation_ReproduceSexual(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	before := len(sim.Organisms())
	orgs := sim.Organisms()
	p1, p2 := orgs[0], orgs[1]

	offspring, err := sim.ReproduceSexual(p1, p2)
	require.NoError(t, err)
	require.NotNil(t, offspring)

	assert.Len(t, sim.Organisms(), before+1)
	assert.True(t, offspring.IsAlive())
	assert.NotNil(t, offspring.Species)
}

// TestSimulation_Kill verifies that Kill removes an organism and updates population size.
func TestSimulation_Kill(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	before := len(sim.Organisms())
	target := sim.Organisms()[0]

	err = sim.Kill(target)
	require.NoError(t, err)

	assert.Len(t, sim.Organisms(), before-1)
	assert.False(t, target.IsAlive())
}

// TestSimulation_KillWhere verifies that KillWhere removes all matching organisms.
func TestSimulation_KillWhere(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	// Kill all organisms born at tick 1 (the initial population)
	count, err := sim.KillWhere(func(org *genetics.Organism) bool {
		return org.Generation == 1
	})
	require.NoError(t, err)
	assert.Equal(t, testOpts().PopSize, count)
	assert.Empty(t, sim.Organisms())
}

// TestSimulation_Observer verifies that observer callbacks fire.
func TestSimulation_Observer(t *testing.T) {
	g := loadTestGenome(t)
	sim, err := NewSimulation(context.Background(), g, testOpts())
	require.NoError(t, err)

	var born, died int
	obs := &countingObserver{bornCb: func() { born++ }, diedCb: func() { died++ }}
	sim.RegisterObserver(obs)

	parent := sim.Organisms()[0]
	_, err = sim.ReproduceAsexual(parent)
	require.NoError(t, err)
	assert.Equal(t, 1, born)

	err = sim.Kill(parent)
	require.NoError(t, err)
	assert.Equal(t, 1, died)
}

// countingObserver is a minimal Observer for testing callbacks.
type countingObserver struct {
	bornCb func()
	diedCb func()
}

func (c *countingObserver) OnOrganismBorn(_ *Simulation, _ *genetics.Organism)  { c.bornCb() }
func (c *countingObserver) OnOrganismDied(_ *Simulation, _ *genetics.Organism)  { c.diedCb() }
func (c *countingObserver) OnSpeciesFormed(_ *Simulation, _ *genetics.Species)  {}
func (c *countingObserver) OnSpeciesExtinct(_ *Simulation, _ *genetics.Species) {}
