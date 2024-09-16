package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

type BestLayoutEntry struct {
	Layout  Layout
	Penalty float64
}

func Optimize(quartadInfo QuartadInfo, layout Layout, user User, debug bool, iterations int, topLayouts int, numSwaps int, parallelism int) {
	initLayout := layout.Duplicate()
	penaltyRules := InitPenaltyRules(user)

	if debug {
		fmt.Println("Initial layout:")
		fmt.Print(initLayout.String())
	}

	runesToKeyPhysicalKeyInfoMap := initLayout.mapRunesToPhysicalKeyInfo()
	initialPenalty, initialResults := CalculatePenalty(quartadInfo.Quartads, initLayout, runesToKeyPhysicalKeyInfoMap, &penaltyRules, debug)
	watermarkPenalty := initialPenalty
	if debug {
		fmt.Println()
		PrintPenaltyResults(initialPenalty, watermarkPenalty, initialResults)
	}

	// Initialize simulated annealing
	sa := NewSimulatedAnnealing(iterations)

	// Initialize best layouts list
	var bestLayouts []BestLayoutEntry
	bestLayouts = append(bestLayouts, BestLayoutEntry{Layout: initLayout.Duplicate(), Penalty: initialPenalty})

	acceptedLayout := initLayout.Duplicate()
	acceptedPenalty := initialPenalty
	acceptedPenaltyResults := initialResults

	// Capture the start time for ETA calculation
	startTime := time.Now()

	start, end := sa.GetSimulationRange()

	// Create channels to handle results
	type CalcPenaltyRunResult struct {
		index   int
		layout  Layout
		penalty float64
		results []KeyPenaltyResult
	}

	penaltyChan := make(chan CalcPenaltyRunResult, parallelism)
	var wg sync.WaitGroup

	// Set up signal handling for graceful shutdown on Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Println("\nReceived interrupt signal, shutting down...")
		// Don't close penaltyChan yet, just exit after ongoing work is done
		os.Exit(1) // Gracefully exit the program         // Gracefully exit the program
	}()

	for i := start; i < end; i += parallelism {
		// Launch parallelism number of goroutines
		for j := 0; j < parallelism && i+j < end; j++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Shuffle the layout and calculate penalty
				currLayout := acceptedLayout.Duplicate()
				currLayout.Shuffle(rand.Intn(numSwaps) + 1)
				runesToKeyPhysicalKeyInfoMap := currLayout.mapRunesToPhysicalKeyInfo()

				// Calculate penalty
				currPenalty, currPenaltyResults := CalculatePenalty(quartadInfo.Quartads, currLayout, runesToKeyPhysicalKeyInfoMap, &penaltyRules, debug)

				// Send the result back via channel
				penaltyChan <- CalcPenaltyRunResult{
					index:   idx,
					layout:  currLayout,
					penalty: currPenalty,
					results: currPenaltyResults,
				}
			}(i + j)
		}

		// Start a goroutine that closes the penalty channel when all work is done
		go func() {
			wg.Wait()          // Wait for all goroutines to finish
			close(penaltyChan) // Close the channel after all goroutines are done
		}()

		for result := range penaltyChan {
			// Decide whether to accept the new layout
			if sa.AcceptTransition(result.penalty-acceptedPenalty, result.index) {
				acceptedLayout = result.layout.Duplicate()
				acceptedPenalty = result.penalty
				acceptedPenaltyResults = result.results
				if acceptedPenalty > watermarkPenalty {
					watermarkPenalty = acceptedPenalty
				}

				PrintProgress(startTime, result.index, end, acceptedLayout, debug, acceptedPenalty, watermarkPenalty, acceptedPenaltyResults)

				// Add the new layout to bestLayouts and maintain top layouts
				bestLayouts = append(bestLayouts, BestLayoutEntry{Layout: result.layout.Duplicate(), Penalty: result.penalty})

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

		if i%100 == 0 {
			PrintProgress(startTime, i, end, acceptedLayout, debug, acceptedPenalty, watermarkPenalty, acceptedPenaltyResults)
		}
	}

	// Print the best layouts found
	fmt.Printf("\nTop %d layouts:\n", topLayouts)
	for i, entry := range bestLayouts {
		fmt.Printf("\nBest layout #%d:\n", i+1)
		fmt.Print(entry.Layout.String())
		fmt.Println()
		finalPenalty, finalResults := CalculatePenalty(quartadInfo.Quartads, entry.Layout, runesToKeyPhysicalKeyInfoMap, &penaltyRules, debug)
		if finalPenalty > watermarkPenalty {
			watermarkPenalty = finalPenalty
		}
		PrintPenaltyResults(finalPenalty, watermarkPenalty, finalResults)
	}
}

func PrintProgress(startTime time.Time, i int, end int, acceptedLayout Layout, debug bool, acceptedPenalty float64, watermarkPenalty float64, acceptedPenaltyResults []KeyPenaltyResult) {
	// Calculate elapsed time and ETA
	elapsed := time.Since(startTime)
	progress := float64(i) / float64(end-1)
	eta := time.Duration(float64(elapsed) * (1.0/progress - 1.0))

	fmt.Println()
	fmt.Println(acceptedLayout.String())
	fmt.Printf("  Iteration %d/%d (%.3g%% complete) | ETA: %s\n", i, end-1, progress*100.0, eta.Round(time.Second))

	if debug {
		fmt.Println()
		PrintPenaltyResults(acceptedPenalty, watermarkPenalty, acceptedPenaltyResults)
	}
}

func PrintPenaltyResults(current, watermark float64, results []KeyPenaltyResult) {
	penaltyPercentage := (current / watermark) * 100.0
	fmt.Printf("  %30s: %s %.3g%%\n", fmt.Sprintf("Layout penalty: %.0f", current), generateProgressBar(penaltyPercentage, 16), penaltyPercentage)
	fmt.Println()
	for _, r := range results {
		percentOfHighest := r.Total / r.WatermarkPenalty * 100.0
		if r.Info.Cost < 0 {
			percentOfHighest = math.Abs(r.WatermarkPenalty-r.Total) / math.Abs(r.WatermarkPenalty) * 100.0
		}
		fmt.Printf("  %30s: %s %.3g%%\n", r.Name, generateProgressBar(percentOfHighest, 16), percentOfHighest)
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
