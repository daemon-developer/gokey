package main

import (
	"fmt"
	"math/rand"
	"sort"
)

type BestLayoutEntry struct {
	Layout  Layout
	Penalty float64
}

func Optimize(quartadInfo QuartadInfo, layout Layout, user User, debug bool, iterations int, topLayouts int, numSwaps int) {
	initLayout := layout.Duplicate()
	penaltyRules := InitPenaltyRules(user)

	if debug {
		fmt.Println("Initial layout:")
		fmt.Print(initLayout.String())
	}

	runesToKeyPhysicalKeyInfoMap := initLayout.mapRunesToPhysicalKeyInfo()
	initialPenalty, initialResults := CalculatePenalty(quartadInfo.Quartads, initLayout, runesToKeyPhysicalKeyInfoMap, penaltyRules, true)
	if debug {
		fmt.Printf("Initial layout penalty: %f\n", initialPenalty)
		for _, r := range initialResults {
			fmt.Printf("  %s penalty %f\n", r.Name, r.Total)
		}
	}

	// Initialize simulated annealing
	sa := NewSimulatedAnnealing(iterations)

	// Initialize best layouts list
	var bestLayouts []BestLayoutEntry
	bestLayouts = append(bestLayouts, BestLayoutEntry{Layout: initLayout.Duplicate(), Penalty: initialPenalty})

	acceptedLayout := initLayout.Duplicate()
	acceptedPenalty := initialPenalty

	start, end := sa.GetSimulationRange()
	for i := start; i < end; i++ {
		if i%500 == 0 {
			fmt.Println(acceptedLayout.String())
		}

		// Create a new layout by shuffling the accepted layout
		currLayout := acceptedLayout.Duplicate()
		currLayout.Shuffle(rand.Intn(numSwaps) + 1)
		runesToKeyPhysicalKeyInfoMap := currLayout.mapRunesToPhysicalKeyInfo()

		// Calculate the penalty for the new layout
		currPenalty, _ := CalculatePenalty(quartadInfo.Quartads, currLayout, runesToKeyPhysicalKeyInfoMap, penaltyRules, false)

		// Decide whether to accept the new layout
		if sa.AcceptTransition(currPenalty-acceptedPenalty, i) {
			if debug {
				fmt.Printf("Iteration %d accepted with penalty %.0f\n", i, currPenalty)
			}

			acceptedLayout = currLayout.Duplicate()
			acceptedPenalty = currPenalty

			// Add the new layout to bestLayouts and maintain top layouts
			bestLayouts = append(bestLayouts, BestLayoutEntry{Layout: currLayout.Duplicate(), Penalty: currPenalty})

			// Sort bestLayouts by penalty (lowest first)
			sort.Slice(bestLayouts, func(i, j int) bool {
				return bestLayouts[i].Penalty < bestLayouts[j].Penalty
			})

			// Keep only the top layouts
			if len(bestLayouts) > topLayouts {
				bestLayouts = bestLayouts[:topLayouts]
			}
		}
	}

	// Print the best layouts found
	fmt.Printf("\nTop %d layouts:\n", topLayouts)
	for i, entry := range bestLayouts {
		fmt.Printf("\nBest layout #%d:\n", i+1)
		fmt.Print(entry.Layout.String())
		finalPenalty, finalResults := CalculatePenalty(quartadInfo.Quartads, entry.Layout, runesToKeyPhysicalKeyInfoMap, penaltyRules, true)
		fmt.Printf("Final layout penalty: %f\n", finalPenalty)
		for _, r := range finalResults {
			fmt.Printf("  %s penalty %f\n", r.Name, r.Total)
		}
	}
}

// insertSortedBestLayouts inserts a new layout into the sorted list of best layouts

func (layout *Layout) Shuffle(numSwaps int) {
	swappableKeys := layout.GetSwappableKeys()

	for i := 0; i < numSwaps; i++ {
		layout.shufflePosition(swappableKeys)
	}
}

func (layout *Layout) shufflePosition(swappableKeys []*Key) {
	if len(swappableKeys) < 2 {
		return // Not enough swappable keys to perform a swap
	}

	// Select two distinct random indices
	i := rand.Intn(len(swappableKeys))
	j := rand.Intn(len(swappableKeys) - 1)
	if j >= i {
		j++
	}

	// Swap the content of the keys
	swapKeys(swappableKeys[i], swappableKeys[j])
}

func swapKeys(a, b *Key) {
	a.UnshiftedRune, b.UnshiftedRune = b.UnshiftedRune, a.UnshiftedRune
	a.UnshiftedIsFree, b.UnshiftedIsFree = b.UnshiftedIsFree, a.UnshiftedIsFree
	a.ShiftedRune, b.ShiftedRune = b.ShiftedRune, a.ShiftedRune
	a.ShiftedIsFree, b.ShiftedIsFree = b.ShiftedIsFree, a.ShiftedIsFree
}
