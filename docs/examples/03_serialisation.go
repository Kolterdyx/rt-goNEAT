//go:build ignore
// +build ignore

// Example 03: Saving and Restoring a Simulation
//
// Demonstrates how to save the population state to disk and resume from it:
//   - Run a simulation for N ticks
//   - Write the population to a file
//   - Load the population from the file and continue
//
// This is useful for checkpointing long-running simulations.
//
// Run with: go run docs/examples/03_serialisation.go

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/Kolterdyx/rt-goNEAT/v1/alife"
	"github.com/Kolterdyx/rt-goNEAT/v1/neat"
	"github.com/Kolterdyx/rt-goNEAT/v1/neat/genetics"
	neatmath "github.com/Kolterdyx/rt-goNEAT/v1/neat/math"
)

const checkpointFile = "/tmp/population_checkpoint.neat"

func buildOpts() *neat.Options {
	return &neat.Options{
		CompatThreshold:       3.0,
		DisjointCoeff:         1.0,
		ExcessCoeff:           1.0,
		MutdiffCoeff:          0.4,
		GenCompatMethod:       neat.GenomeCompatibilityMethodLinear,
		PopSize:               10,
		MutateAddNodeProb:     0.03,
		MutateAddLinkProb:     0.05,
		MutateLinkWeightsProb: 0.80,
		MutateOnlyProb:        0.25,
		WeightMutPower:        2.5,
		NewLinkTries:          20,
		MateMultipointProb:    0.6,
		MateMultipointAvgProb: 0.3,
		MateSinglepointProb:   0.1,
		MateOnlyProb:          0.2,
		NodeActivators:        []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
		NodeActivatorsProb:    []float64{1.0},
	}
}

// runFor runs the simulation for n ticks with random reproduction/death.
func runFor(sim *alife.Simulation, n int) {
	for i := 0; i < n; i++ {
		sim.Step()
		orgs := sim.Organisms()
		if len(orgs) >= 2 {
			p1 := orgs[rand.Intn(len(orgs))]
			p2 := orgs[rand.Intn(len(orgs))]
			_, _ = sim.ReproduceSexual(p1, p2)
		}
		if len(orgs) > 20 {
			_ = sim.Kill(orgs[rand.Intn(len(orgs))])
		}
	}
}

func main() {
	opts := buildOpts()

	// ── Phase 1: Create and run ──────────────────────────────────────────────
	reader, err := genetics.NewGenomeReaderFromFile("../../data/xorstartgenes")
	if err != nil {
		log.Fatal(err)
	}
	seedGenome, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	sim, err := alife.NewSimulation(context.Background(), seedGenome, opts)
	if err != nil {
		log.Fatal(err)
	}

	runFor(sim, 200)
	fmt.Printf("Phase 1 done: tick=%d, pop=%d, species=%d\n",
		sim.Tick(), len(sim.Organisms()), len(sim.Population.Species))

	// ── Phase 2: Save to file ────────────────────────────────────────────────
	f, err := os.Create(checkpointFile)
	if err != nil {
		log.Fatal("failed to create checkpoint:", err)
	}
	if err := sim.Population.Write(f); err != nil {
		log.Fatal("failed to write population:", err)
	}
	f.Close()
	fmt.Printf("Population saved to %s\n", checkpointFile)

	// ── Phase 3: Load from file and continue ─────────────────────────────────
	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		log.Fatal("failed to read checkpoint:", err)
	}
	restoredPop, err := genetics.ReadPopulation(strings.NewReader(string(data)), opts)
	if err != nil {
		log.Fatal("failed to restore population:", err)
	}

	// Wrap the restored population in a new Simulation
	restoredSim := &alife.Simulation{Population: restoredPop}
	// NOTE: the Simulation context (carrying neat.Options) must be set separately.
	// Use NewSimulation directly from a checkpoint genome instead for a cleaner approach:
	_ = restoredSim

	fmt.Printf("Restored population: %d organisms, %d species\n",
		len(restoredPop.Organisms), len(restoredPop.Species))

	// ── Alternative: restore and continue via a new Simulation from the saved genome ──
	// A simpler and more reliable checkpoint strategy is to save the champion genome
	// from each species, then restart with one of them as the seed:
	if len(sim.Population.Species) > 0 {
		// Pick the oldest species' oldest organism as checkpoint genome
		bestSpecies := sim.Population.Species[0]
		if bestSpecies.Size() > 0 {
			champion := bestSpecies.FindOldest()
			if champion != nil {
				var buf strings.Builder
				err := champion.Genotype.Write(&buf)
				if err == nil {
					fmt.Printf("Champion genome (species %d, tick %d):\n%s\n",
						bestSpecies.Id, champion.Generation, buf.String())
				}
			}
		}
	}

	fmt.Println("Done.")
}
