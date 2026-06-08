//go:build ignore
// +build ignore

// Example 02: Energy-Based Lifecycle
//
// Demonstrates a more realistic ALife pattern:
//   - Each organism has an energy level stored in its Data field
//   - Organisms lose energy each tick (metabolic cost + cost per network node)
//   - Organisms gain energy from "food" that spawns at random positions
//   - When energy exceeds a threshold, the organism reproduces asexually
//   - When energy reaches 0, the organism dies
//   - Observers log births, deaths, and species events to stdout
//
// This pattern shows how to encode simulation state in organism.Data,
// how to use observers for event-driven logic, and how network complexity
// has a runtime cost (more nodes = higher metabolism).
//
// Run with: go run docs/examples/02_energy_lifecycle.go

package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/Kolterdyx/rt-goNEAT/alife"
	"github.com/Kolterdyx/rt-goNEAT/neat"
	"github.com/Kolterdyx/rt-goNEAT/neat/genetics"
	neatmath "github.com/Kolterdyx/rt-goNEAT/neat/math"
)

// ── Organism state ──────────────────────────────────────────────────────────

// State holds the ALife data for one organism.
type State struct {
	Energy float64
	X, Y   float64 // position in [0, 1]^2
	DX, DY float64 // velocity
}

func newState() *State {
	return &State{
		Energy: 50.0,
		X:      rand.Float64(),
		Y:      rand.Float64(),
	}
}

func getState(org *genetics.Organism) *State {
	if org.Data == nil {
		return nil
	}
	return org.Data.Value.(*State)
}

// ── Food ────────────────────────────────────────────────────────────────────

type Food struct {
	X, Y   float64
	Energy float64
}

func randomFood() Food {
	return Food{X: rand.Float64(), Y: rand.Float64(), Energy: 30.0}
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx, dy := x1-x2, y1-y2
	return math.Sqrt(dx*dx + dy*dy)
}

// ── Observer ────────────────────────────────────────────────────────────────

type WorldObserver struct {
	births, deaths int
}

func (w *WorldObserver) OnOrganismBorn(sim *alife.Simulation, org *genetics.Organism) {
	w.births++
}

func (w *WorldObserver) OnOrganismDied(sim *alife.Simulation, org *genetics.Organism) {
	w.deaths++
}

func (w *WorldObserver) OnSpeciesFormed(sim *alife.Simulation, sp *genetics.Species) {
	fmt.Printf("[tick %4d] new species #%d formed\n", sim.Tick(), sp.Id)
}

func (w *WorldObserver) OnSpeciesExtinct(sim *alife.Simulation, sp *genetics.Species) {
	fmt.Printf("[tick %4d] species #%d went extinct\n", sim.Tick(), sp.Id)
}

// ── Main ────────────────────────────────────────────────────────────────────

func main() {
	// Load seed genome (2 inputs: food-dx, food-dy; 2 outputs: thrust-x, thrust-y)
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
		PopSize:               30,
		MutateAddNodeProb:     0.02,
		MutateAddLinkProb:     0.05,
		MutateLinkWeightsProb: 0.80,
		MutateOnlyProb:        0.30,
		WeightMutPower:        2.0,
		NewLinkTries:          20,
		MateMultipointProb:    0.6,
		MateMultipointAvgProb: 0.3,
		MateSinglepointProb:   0.1,
		MateOnlyProb:          0.2,
		InterspeciesMateRate:  0.001,
		NodeActivators:        []neatmath.NodeActivationType{neatmath.SigmoidSteepenedActivation},
		NodeActivatorsProb:    []float64{1.0},
	}

	sim, err := alife.NewSimulation(context.Background(), seedGenome, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Attach energy state to every initial organism
	for _, org := range sim.Organisms() {
		org.Data = &genetics.OrganismData{Value: newState()}
	}

	obs := &WorldObserver{}
	sim.RegisterObserver(obs)

	// Spawn initial food
	foods := make([]Food, 20)
	for i := range foods {
		foods[i] = randomFood()
	}

	const maxTicks = 2000
	const reproThreshold = 120.0 // energy needed to reproduce
	const foodEatRadius = 0.05
	const metabolicBase = 0.2     // energy cost per tick
	const metabolicPerNode = 0.01 // complexity tax

	for sim.Tick() < maxTicks {
		sim.Step()
		tick := sim.Tick()

		// Spawn new food every 10 ticks
		if tick%10 == 0 && len(foods) < 50 {
			foods = append(foods, randomFood())
		}

		orgs := sim.Organisms()
		var toKill []*genetics.Organism
		var toReproduce []*genetics.Organism

		for _, org := range orgs {
			state := getState(org)
			if state == nil {
				continue
			}

			// ── Activate network to get movement direction ──
			net, err := org.Phenotype()
			if err != nil {
				continue
			}
			// Find nearest food for input
			nearestDX, nearestDY := 0.5, 0.5
			bestDist := math.MaxFloat64
			bestFoodIdx := -1
			for i, f := range foods {
				d := dist(state.X, state.Y, f.X, f.Y)
				if d < bestDist {
					bestDist = d
					nearestDX = f.X - state.X
					nearestDY = f.Y - state.Y
					bestFoodIdx = i
				}
			}
			_ = net.LoadSensors([]float64{nearestDX, nearestDY})
			net.Activate()
			outputs := net.ReadOutputs()

			// Move organism based on network output
			if len(outputs) >= 2 {
				state.DX = (outputs[0] - 0.5) * 0.05
				state.DY = (outputs[1] - 0.5) * 0.05
			}
			state.X = math.Mod(state.X+state.DX+1.0, 1.0)
			state.Y = math.Mod(state.Y+state.DY+1.0, 1.0)

			// Eat food if close enough
			if bestFoodIdx >= 0 && bestDist < foodEatRadius {
				state.Energy += foods[bestFoodIdx].Energy
				// Replace eaten food with a new one
				foods[bestFoodIdx] = randomFood()
			}

			// Metabolic cost: base + complexity tax
			nodes := len(org.Genotype.Nodes)
			state.Energy -= metabolicBase + float64(nodes)*metabolicPerNode

			// Mark for death or reproduction
			if state.Energy <= 0 {
				toKill = append(toKill, org)
			} else if state.Energy >= reproThreshold {
				toReproduce = append(toReproduce, org)
			}
		}

		// Kill starved organisms
		for _, org := range toKill {
			_ = sim.Kill(org)
		}

		// Reproduce thriving organisms
		for _, org := range toReproduce {
			if !org.IsAlive() {
				continue // was killed this tick
			}
			state := getState(org)
			// Split energy between parent and offspring
			state.Energy /= 2.0
			offspring, err := sim.ReproduceAsexual(org)
			if err != nil {
				continue
			}
			offspring.Data = &genetics.OrganismData{Value: &State{
				Energy: state.Energy,
				X:      state.X + (rand.Float64()-0.5)*0.1,
				Y:      state.Y + (rand.Float64()-0.5)*0.1,
			}}
		}

		// Print stats every 100 ticks
		if tick%100 == 0 {
			fmt.Printf("Tick %4d | pop=%3d | species=%2d | food=%2d | births=%d deaths=%d\n",
				tick, len(sim.Organisms()), len(sim.Population.Species),
				len(foods), obs.births, obs.deaths)
		}

		// End early if population dies out
		if len(sim.Organisms()) == 0 {
			fmt.Println("Population extinct!")
			break
		}
	}

	fmt.Printf("\nFinal state: pop=%d, species=%d, total births=%d, total deaths=%d\n",
		len(sim.Organisms()), len(sim.Population.Species), obs.births, obs.deaths)
}
