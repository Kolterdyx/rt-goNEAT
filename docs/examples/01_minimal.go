//go:build ignore
// +build ignore

// Example 01: Minimal ALife Simulation
//
// Demonstrates the smallest possible rt-goNEAT program:
//   - Load a seed genome
//   - Create a simulation
//   - Run a loop that reproduces and kills organisms
//   - Activate each organism's network each tick
//
// Run with: go run docs/examples/01_minimal.go

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
	// ── 1. Load seed genome ─────────────────────────────────────────────────
	// The XOR start genome has 2 inputs + 1 bias → 1 output.
	reader, err := genetics.NewGenomeReaderFromFile("../../data/xorstartgenes")
	if err != nil {
		log.Fatal("failed to open genome:", err)
	}
	seedGenome, err := reader.Read()
	if err != nil {
		log.Fatal("failed to read genome:", err)
	}

	// ── 2. Configure ────────────────────────────────────────────────────────
	opts := &neat.Options{
		// Speciation
		CompatThreshold: 3.0,
		DisjointCoeff:   1.0,
		ExcessCoeff:     1.0,
		MutdiffCoeff:    0.4,
		GenCompatMethod: neat.GenomeCompatibilityMethodLinear,

		// Seed population size
		PopSize: 20,

		// Mutation probabilities
		MutateAddNodeProb:     0.03,
		MutateAddLinkProb:     0.05,
		MutateConnectSensors:  0.10,
		MutateLinkWeightsProb: 0.80,
		MutateOnlyProb:        0.25,
		WeightMutPower:        2.5,
		NewLinkTries:          20,

		// Mating probabilities
		MateMultipointProb:    0.60,
		MateMultipointAvgProb: 0.30,
		MateSinglepointProb:   0.10,
		MateOnlyProb:          0.20,
		InterspeciesMateRate:  0.001,

		// Activation functions for new hidden neurons
		NodeActivators:     []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
		NodeActivatorsProb: []float64{1.0},
	}

	// ── 3. Create simulation ─────────────────────────────────────────────────
	sim, err := alife.NewSimulation(context.Background(), seedGenome, opts)
	if err != nil {
		log.Fatal("failed to create simulation:", err)
	}

	fmt.Printf("Simulation started: %d organisms, %d species\n",
		len(sim.Organisms()), len(sim.Population.Species))

	// ── 4. Run simulation loop ───────────────────────────────────────────────
	const maxTicks = 500
	const maxPop = 40
	const minPop = 10

	for sim.Tick() < maxTicks {
		sim.Step()
		orgs := sim.Organisms()

		// Activate each organism's network with random inputs.
		// In a real ALife simulation, inputs would come from the environment
		// (e.g., nearby food, distance to wall, current speed).
		for _, org := range orgs {
			net, err := org.Phenotype()
			if err != nil {
				continue
			}
			// XOR start genome: 2 sensor inputs + 1 bias = 3 sensor values
			// LoadSensors only touches sensor (non-bias) nodes — 2 inputs here
			_ = net.LoadSensors([]float64{rand.Float64(), rand.Float64()})
			activated, _ := net.Activate()
			if !activated {
				continue
			}
			outputs := net.ReadOutputs()
			_ = outputs // use outputs to drive simulation behaviour
		}

		// Reproduce: randomly pick parents and create an offspring.
		// In a real simulation, reproduction would depend on organism state:
		// e.g., when an organism accumulates enough energy.
		if len(orgs) >= 2 && len(orgs) < maxPop {
			p1 := orgs[rand.Intn(len(orgs))]
			p2 := orgs[rand.Intn(len(orgs))]
			if _, err := sim.ReproduceSexual(p1, p2); err != nil {
				log.Println("reproduce error:", err)
			}
		}

		// Kill random old organism to keep population bounded.
		if len(orgs) > minPop {
			target := orgs[rand.Intn(len(orgs))]
			if err := sim.Kill(target); err != nil {
				log.Println("kill error:", err)
			}
		}

		// Print a progress line every 50 ticks.
		if sim.Tick()%50 == 0 {
			fmt.Printf("Tick %4d | pop=%3d | species=%d\n",
				sim.Tick(), len(sim.Organisms()), len(sim.Population.Species))
		}
	}

	fmt.Println("Simulation complete.")
}
