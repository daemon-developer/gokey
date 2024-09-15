package main

import (
	"fmt"
	"math/rand"
	"sort"
	"unicode"
)

type BestLayoutEntry struct {
	Layout  Layout
	Penalty float64
}

func Optimize(quartadInfo QuartadInfo, layout Layout, user User, debug bool, topLayouts int, numSwaps int) {
	initLayout := CreateInitialLayout(quartadInfo, layout, user)
	runesToKeyPhysicalKeyInfoMap := initLayout.mapRunesToPhysicalKeyInfo()
	penaltyRules := InitPenaltyRules(user)

	if debug {
		fmt.Println("Initial layout:")
		fmt.Print(initLayout.String())
	}

	initialPenalty, initialResults := CalculatePenalty(quartadInfo.Quartads, initLayout, runesToKeyPhysicalKeyInfoMap, penaltyRules, true)
	if debug {
		fmt.Printf("Initial layout penalty: %f\n", initialPenalty)
		for _, r := range initialResults {
			fmt.Printf("  %s penalty %f\n", r.Name, r.Total)
		}
	}

	// Initialize simulated annealing
	sa := NewSimulatedAnnealing()

	// Initialize best layouts list
	var bestLayouts []BestLayoutEntry
	bestLayouts = append(bestLayouts, BestLayoutEntry{Layout: initLayout.Duplicate(), Penalty: initialPenalty})

	acceptedLayout := initLayout.Duplicate()
	acceptedPenalty := initialPenalty

	start, end := sa.GetSimulationRange()
	for i := start; i < end; i++ {
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
	a.MustUseUnshiftedRune, b.MustUseUnshiftedRune = b.MustUseUnshiftedRune, a.MustUseUnshiftedRune
	a.FreeToPlaceUnshifted, b.FreeToPlaceUnshifted = b.FreeToPlaceUnshifted, a.FreeToPlaceUnshifted
	a.MustUseShiftedRune, b.MustUseShiftedRune = b.MustUseShiftedRune, a.MustUseShiftedRune
	a.FreeToPlaceShifted, b.FreeToPlaceShifted = b.FreeToPlaceShifted, a.FreeToPlaceShifted
}

func CreateInitialLayout(quartadInfo QuartadInfo, layout Layout, user User) Layout {
	l := deepCopyLayout(layout)

	assignRunesToKeys(l, quartadInfo.RunesToPlace, user.Locale)

	return l
}

func assignRunesToKeys(layout Layout, validRunes []rune, locale Locale) {
	orderedKeys := getOrderedKeysByCost(layout)
	var i int

	alreadyAssigned := make(map[rune]bool)

	for _, key := range orderedKeys {
		if i >= len(validRunes) {
			break
		}

		// Get the rune to assign
		runeToAssign := validRunes[i]
		if runeToAssign == ' ' {
			fmt.Println("Check this")
		}

		// Check if it's already assigned
		if alreadyAssigned[runeToAssign] {
			i++
			continue
		}

		if unicode.IsLetter(runeToAssign) {
			if key.FreeToPlaceUnshifted == false {
				continue
			}
			lowerAlpha := unicode.ToLower(runeToAssign)
			if !alreadyAssigned[lowerAlpha] {
				// If it's a letter, assign lower and upper case
				key.MustUseUnshiftedRune = lowerAlpha
				key.MustUseShiftedRune = unicode.ToUpper(runeToAssign)
				key.FreeToPlaceUnshifted = false
				key.FreeToPlaceShifted = false
				alreadyAssigned[lowerAlpha] = true
			}
		} else {
			// Handle symbol assignment
			if layout.SupportsOverrides {
				if key.FreeToPlaceUnshifted {
					key.MustUseUnshiftedRune = runeToAssign
					key.FreeToPlaceUnshifted = false
					alreadyAssigned[runeToAssign] = true
				} else {
					key.MustUseShiftedRune = runeToAssign
					key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
				}
			} else {
				if key.FreeToPlaceUnshifted == false {
					continue
				}
				// Use locale-based symbol mapping if overrides aren't supported
				if shiftedRune, exists := locale.unshiftedToShifted[runeToAssign]; exists {
					key.MustUseUnshiftedRune = runeToAssign
					key.MustUseShiftedRune = shiftedRune
					key.FreeToPlaceUnshifted = false
					key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[shiftedRune] = true
				} else if unshiftedRune, exists := locale.shiftedToUnshifted[runeToAssign]; exists {
					key.MustUseUnshiftedRune = unshiftedRune
					key.MustUseShiftedRune = runeToAssign
					key.FreeToPlaceUnshifted = false
					key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[unshiftedRune] = true
				} else {
					// Not actually symbols so will be things like ENTER or backspace
					key.MustUseUnshiftedRune = runeToAssign
					key.MustUseShiftedRune = runeToAssign
					key.FreeToPlaceUnshifted = false
					key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
				}
			}
		}

		i++
	}
}

func getOrderedKeysByCost(layout Layout) []*Key {
	var keys []*Key

	for r, _ := range layout.Left.Rows {
		for c, _ := range layout.Left.Rows[r] {
			keys = append(keys, &layout.Left.Rows[r][c])
		}
	}
	for r, _ := range layout.Right.Rows {
		for c, _ := range layout.Right.Rows[r] {
			keys = append(keys, &layout.Right.Rows[r][c])
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Cost < keys[j].Cost
	})

	return keys
}
