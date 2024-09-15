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
	l := layout.Duplicate()
	assignRunesToKeys(l, quartadInfo.RunesToPlace, user)
	return l
}

func assignRunesToKeys(layout Layout, validRunes []rune, user User) {
	orderedKeyInfos := getOrderedKeysByCost(layout)
	var i int

	alreadyAssigned := make(map[rune]bool)

	for _, keyInfo := range orderedKeyInfos {
		if i >= len(validRunes) {
			break
		}

		// Get the rune to assign
		runeToAssign := validRunes[i]

		// Check if it's already assigned
		if alreadyAssigned[runeToAssign] {
			i++
			continue
		}

		if unicode.IsLetter(runeToAssign) {
			if keyInfo.key.FreeToPlaceUnshifted == false {
				continue
			}
			lowerAlpha := unicode.ToLower(runeToAssign)
			if alreadyAssigned[lowerAlpha] == false {
				// If it's a letter, assign lower and upper case
				keyInfo.key.MustUseUnshiftedRune = lowerAlpha
				keyInfo.key.MustUseShiftedRune = unicode.ToUpper(runeToAssign)
				keyInfo.key.FreeToPlaceUnshifted = false
				keyInfo.key.FreeToPlaceShifted = false
				alreadyAssigned[lowerAlpha] = true
			}
		} else {
			// Handle symbol assignment
			if layout.SupportsOverrides {
				if keyInfo.key.FreeToPlaceUnshifted {
					keyInfo.key.MustUseUnshiftedRune = runeToAssign
					keyInfo.key.FreeToPlaceUnshifted = false
					alreadyAssigned[runeToAssign] = true
				} else {
					keyInfo.key.MustUseShiftedRune = runeToAssign
					keyInfo.key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
				}
			} else {
				if keyInfo.key.FreeToPlaceUnshifted == false {
					continue
				}
				// Use locale-based symbol mapping if overrides aren't supported
				if shiftedRune, exists := user.Locale.unshiftedToShifted[runeToAssign]; exists {
					keyInfo.key.MustUseUnshiftedRune = runeToAssign
					keyInfo.key.MustUseShiftedRune = shiftedRune
					keyInfo.key.FreeToPlaceUnshifted = false
					keyInfo.key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[shiftedRune] = true
				} else if unshiftedRune, exists := user.Locale.shiftedToUnshifted[runeToAssign]; exists {
					skip := false
					for _, r := range user.Required {
						if unshiftedRune == r {
							// Already handled
							skip = true
							break
						}
					}
					if !skip {
						for _, r := range layout.EssentialRunes {
							if unshiftedRune == r {
								// Already handled
								skip = true
								break
							}
						}
					}
					if skip {
						i++
						continue
					}
					keyInfo.key.MustUseUnshiftedRune = unshiftedRune
					keyInfo.key.MustUseShiftedRune = runeToAssign
					keyInfo.key.FreeToPlaceUnshifted = false
					keyInfo.key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[unshiftedRune] = true
				} else {
					// Not actually symbols so will be things like ENTER or backspace
					keyInfo.key.MustUseUnshiftedRune = runeToAssign
					keyInfo.key.MustUseShiftedRune = runeToAssign
					keyInfo.key.FreeToPlaceUnshifted = false
					keyInfo.key.FreeToPlaceShifted = false
					alreadyAssigned[runeToAssign] = true
				}
			}
		}

		fmt.Printf("Assigned '%c' to a key\n", runeToAssign)

		i++
	}
}

func getOrderedKeysByCost(layout Layout) []*KeyPhysicalInfo {
	var keyInfos []*KeyPhysicalInfo

	for r := range layout.Left.Rows {
		for c := range layout.Left.Rows[r] {
			keyInfos = append(keyInfos, &layout.Left.Rows[r][c])
		}
	}
	for r := range layout.Right.Rows {
		for c := range layout.Right.Rows[r] {
			keyInfos = append(keyInfos, &layout.Right.Rows[r][c])
		}
	}

	sort.Slice(keyInfos, func(i, j int) bool {
		return keyInfos[i].cost < keyInfos[j].cost
	})

	return keyInfos
}
