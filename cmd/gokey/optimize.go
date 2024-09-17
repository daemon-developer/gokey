package main

import (
	"fmt"
	"math"
	"time"

	"atomicgo.dev/cursor"
)

type BestLayoutEntry struct {
	Layout  Layout
	Penalty float64
}

func Optimize(quartadInfo QuartadInfo, layout Layout, user User, iterations int, numSwaps int) {
	// Capture the start time for ETA calculation
	startTime := time.Now()

	initLayout := layout.Duplicate()
	penaltyRules := InitPenaltyRules(user)
	outputRows := len(layout.Left.Rows) + 8
	if optDebug > 1 {
		outputRows += len(penaltyRules) + 1
	}

	if optDebug > 0 {
		fmt.Println("Initial layout:")
		fmt.Print(initLayout.String())
	}

	runesToKeyPhysicalKeyInfoMap := initLayout.mapRunesToPhysicalKeyInfo()
	initialPenalty, initialResults := CalculatePenalty(quartadInfo.Quartads, initLayout, runesToKeyPhysicalKeyInfoMap, &penaltyRules)
	watermarkPenalty := initialPenalty
	if watermarkPenalty > user.StartingPenaltyWatermark {
		watermarkPenalty = user.StartingPenaltyWatermark
	}
	PrintProgress(startTime, 0, 1, initLayout, 1.0, 1.0, initialResults, nil)

	// Initialize simulated annealing
	sa := NewSimulatedAnnealing(iterations)

	// Initialize best layouts list
	var bestLayout BestLayoutEntry
	bestLayout = BestLayoutEntry{Layout: initLayout.Duplicate(), Penalty: initialPenalty}

	acceptedLayout := initLayout.Duplicate()
	acceptedPenalty := initialPenalty
	acceptedPenaltyResults := initialResults

	start, end := sa.GetSimulationRange()
	for i := start; i < end; i++ {
		if i%100 == 0 {
			cursor.StartOfLineUp(outputRows)
			PrintProgress(startTime, i, end, acceptedLayout, acceptedPenalty, watermarkPenalty, acceptedPenaltyResults, &bestLayout)
		}

		// Create a new layout by shuffling the accepted layout
		currLayout := acceptedLayout.Duplicate()
		currLayout.Shuffle(r.Intn(numSwaps) + 1)
		runesToKeyPhysicalKeyInfoMap := currLayout.mapRunesToPhysicalKeyInfo()

		// Calculate the penalty for the new layout
		currPenalty, currPenaltyResults := CalculatePenalty(quartadInfo.Quartads, currLayout, runesToKeyPhysicalKeyInfoMap, &penaltyRules)

		// Check if this is the best layout so far
		if currPenalty < bestLayout.Penalty {
			bestLayout = BestLayoutEntry{Layout: currLayout.Duplicate(), Penalty: currPenalty}
		}

		// Decide whether to accept the new layout
		if sa.AcceptTransition(currPenalty-acceptedPenalty, i) {
			acceptedLayout = currLayout.Duplicate()
			acceptedPenalty = currPenalty
			acceptedPenaltyResults = currPenaltyResults
			if acceptedPenalty > watermarkPenalty {
				watermarkPenalty = acceptedPenalty
				if watermarkPenalty > user.StartingPenaltyWatermark {
					watermarkPenalty = user.StartingPenaltyWatermark
				}
			}

			cursor.StartOfLineUp(outputRows)
			PrintProgress(startTime, i, end, acceptedLayout, acceptedPenalty, watermarkPenalty, acceptedPenaltyResults, &bestLayout)
		}
	}

	// Print the best layouts found
	fmt.Println("\nBest layout:")
	runesToKeyPhysicalKeyInfoMap = bestLayout.Layout.mapRunesToPhysicalKeyInfo()
	finalPenalty, finalResults := CalculatePenalty(quartadInfo.Quartads, bestLayout.Layout, runesToKeyPhysicalKeyInfoMap, &penaltyRules)
	if finalPenalty > watermarkPenalty {
		watermarkPenalty = finalPenalty
	}
	PrintProgress(startTime, end, end, bestLayout.Layout, finalPenalty, watermarkPenalty, finalResults, &bestLayout)
}

func PrintProgress(startTime time.Time, i int, end int, acceptedLayout Layout, acceptedPenalty float64, watermarkPenalty float64, acceptedPenaltyResults []KeyPenaltyResult, bestLayout *BestLayoutEntry) {
	// Calculate elapsed time and ETA
	elapsed := time.Since(startTime)
	progress := float64(i) / float64(end-1)
	etaDuration := time.Duration(float64(elapsed) * (1.0/progress - 1.0))
	etaTime := time.Now().Add(etaDuration)

	fmt.Println()
	fmt.Println(acceptedLayout.String())

	printPenaltyResults(acceptedPenalty, watermarkPenalty, acceptedPenaltyResults, bestLayout)
	fmt.Println()
	cursor.ClearLine()
	fmt.Printf("  Iteration %d/%d (%.3g%% complete) | ETA: %s (%s)\n",
		i, end-1, progress*100.0,
		etaTime.Format("15:04:05"),
		etaDuration.Round(time.Second))

}

func printPenaltyResults(current, watermark float64, results []KeyPenaltyResult, bestLayout *BestLayoutEntry) {
	penaltyPercentage := (current / watermark) * 100.0
	cursor.ClearLine()
	fmt.Printf("  %30s: %s %.3g%%\n", fmt.Sprintf("Layout penalty: %.0f", current), generateProgressBar(penaltyPercentage, 16), penaltyPercentage)
	if bestLayout != nil {
		penaltyPercentage = (bestLayout.Penalty / watermark) * 100.0
		cursor.ClearLine()
		fmt.Printf("  %30s: %s %.3g%%\n", fmt.Sprintf("Best layout penalty: %.0f", bestLayout.Penalty), generateProgressBar(penaltyPercentage, 16), penaltyPercentage)
	} else {
		cursor.ClearLine()
		fmt.Println()
	}
	if optDebug > 0 {
		fmt.Println()
		for _, r := range results {
			percentOfHighest := r.Total / r.WatermarkPenalty * 100.0
			if r.Info.Cost < 0 {
				percentOfHighest = math.Abs(r.WatermarkPenalty-r.Total) / math.Abs(r.WatermarkPenalty) * 100.0
			}
			cursor.ClearLine()
			fmt.Printf("  %30s: %s %.3g%% %g\n", r.Name, generateProgressBar(percentOfHighest, 16), percentOfHighest, r.Total)
		}
	}
}

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
	i := r.Intn(len(swappableKeys))
	j := r.Intn(len(swappableKeys) - 1)
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
