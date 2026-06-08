//go:build ignore
// +build ignore

// Example 04: Parallel Network Activation
//
// Demonstrates how to activate thousands of organism networks concurrently
// using a worker pool, while keeping reproduction/death sequential.
//
// The key insight: network activation is read-only per organism (each organism
// has its own independent network). Reproduction and death modify the population
// and must be done on one goroutine (or with careful locking).
//
// Pattern:
//  1. Snapshot the organism list (sim.Organisms())
//  2. Activate all networks in parallel via a worker pool
//  3. Collect results (e.g., outputs → scores)
//  4. Back on main goroutine: reproduce/kill based on scores
//
// Run with: go run docs/examples/04_parallel_activation.go

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"

	"github.com/Kolterdyx/rt-goNEAT/v1/alife"
	"github.com/Kolterdyx/rt-goNEAT/v1/neat"
	"github.com/Kolterdyx/rt-goNEAT/v1/neat/genetics"
	neatmath "github.com/Kolterdyx/rt-goNEAT/v1/neat/math"
)

// activationResult holds one organism's network outputs.
type activationResult struct {
	org     *genetics.Organism
	outputs []float64
	score   float64
}

// activateAll runs all organisms' networks in parallel and returns results.
func activateAll(orgs []*genetics.Organism, workers int) []activationResult {
	type job struct {
		idx int
		org *genetics.Organism
	}

	jobs := make(chan job, len(orgs))
	results := make([]activationResult, len(orgs))
	var wg sync.WaitGroup

	// Spawn workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				net, err := j.org.Phenotype()
				if err != nil {
					continue
				}
				inputs := []float64{rand.Float64(), rand.Float64()}
				_ = net.LoadSensors(inputs)
				activated, _ := net.Activate()
				if !activated {
					continue
				}
				outputs := net.ReadOutputs()
				score := 0.0
				for _, o := range outputs {
					score += o
				}
				results[j.idx] = activationResult{
					org:     j.org,
					outputs: outputs,
					score:   score,
				}
			}
		}()
	}

	// Submit jobs
	for i, org := range orgs {
		jobs <- job{idx: i, org: org}
	}
	close(jobs)
	wg.Wait()

	return results
}

func main() {
	reader, err := genetics.NewGenomeReaderFromFile("../../data/xorstartgenes")
	if err != nil {
		log.Fatal(err)
	}
	seedGenome, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	opts := &neat.Options{
		CompatThreshold:       3.0,
		DisjointCoeff:         1.0,
		ExcessCoeff:           1.0,
		MutdiffCoeff:          0.4,
		GenCompatMethod:       neat.GenomeCompatibilityMethodLinear,
		PopSize:               200, // larger population to demonstrate parallelism
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

	sim, err := alife.NewSimulation(context.Background(), seedGenome, opts)
	if err != nil {
		log.Fatal(err)
	}

	workers := runtime.NumCPU()
	fmt.Printf("Running with %d workers on a population of %d\n",
		workers, len(sim.Organisms()))

	for sim.Tick() < 100 {
		sim.Step()

		// ── Step 1: Snapshot ────────────────────────────────────────────────
		orgs := sim.Organisms()

		// ── Step 2: Parallel activation ─────────────────────────────────────
		results := activateAll(orgs, workers)

		// ── Step 3: Sequential reproduction/death ───────────────────────────
		// Sort could be done here; for simplicity we just use random selection
		// with a bias toward higher-scoring organisms.
		for _, res := range results {
			if res.org == nil || !res.org.IsAlive() {
				continue
			}
			// Score > 0.7: reproduce
			if res.score > 0.7 && len(orgs) < 300 {
				_, _ = sim.ReproduceAsexual(res.org)
			}
			// Score < 0.1: die
			if res.score < 0.1 && len(sim.Organisms()) > 50 {
				_ = sim.Kill(res.org)
			}
		}

		if sim.Tick()%10 == 0 {
			fmt.Printf("Tick %3d | pop=%4d | species=%2d\n",
				sim.Tick(), len(sim.Organisms()), len(sim.Population.Species))
		}
	}

	fmt.Println("Done.")
}
