package main

import (
	"fmt"
	"sort"
	"unicode"
)

func Optimize(quartadInfo QuartadInfo, layout Layout, user User) {
	initLayout := CreateInitialLayout(quartadInfo, layout, user)
	runesToKeyPhysicalKeyInfoMap := initLayout.mapRunesToPhysicalKeyInfo()

	penaltyRules := InitPenaltyRules(user)

	fmt.Print(initLayout.String())

	p, res := CalculatePenalty(quartadInfo.Quartads, initLayout, runesToKeyPhysicalKeyInfoMap, penaltyRules, true)

	fmt.Printf("Initial layout penalty: %f\n", p)
	for _, r := range res {
		fmt.Printf("  %s penalty %f\n", r.Name, r.Total)
	}
}

func CreateInitialLayout(quartadInfo QuartadInfo, layout Layout, user User) Layout {
	l := DeepCopyLayout(layout)

	assignRunesToKeys(l, quartadInfo.ValidRunes, user.Locale)

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

		// Check if it's already assigned
		if alreadyAssigned[runeToAssign] {
			i++
			continue
		}

		if unicode.IsLetter(runeToAssign) {
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
