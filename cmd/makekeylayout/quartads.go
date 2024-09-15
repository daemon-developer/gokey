package main

import (
	"fmt"
	"os"
	"unicode"
)

type QuartadList map[string]int

type QuartadInfo struct {
	Quartads            QuartadList
	RunesToPlace        []rune
	RunesNotBeingPlaced []rune
}

func isTypeableRune(r rune) bool {
	// Check for common typeable characters
	if unicode.IsPrint(r) || r == '\t' || r == '\n' {
		return true
	}

	// Exclude control characters and anything beyond ASCII
	return false
}

func isValidRune(r rune, validRunes map[rune]int) bool {
	_, ok := validRunes[r]
	return ok
}

func PrepareQuartadList(s string, user User) QuartadInfo {
	layout := user.Layout
	foundRunes := make(map[rune]int)
	runesNeedingPlacing := make(map[rune]int)
	quartads := make(QuartadList)
	runes := []rune(s)
	n := len(runes)

	// Count the frequency of all the runes
	for i := 0; i < n; i++ {
		if isTypeableRune(runes[i]) {
			foundRunes[runes[i]]++
		}
	}

	// Create a map to hold valid runes
	validRunes := make(map[rune]int)

	// Ensure essential runes are included
	for _, r := range layout.EssentialRunes {
		validRunes[r] = foundRunes[r]
	}

	// The essential from the layout are essential meet the needs of hardcoded
	// keys on that layout. We now also need to add any the user has asked for
	reqRunes := []rune(user.Required)
	for _, r := range reqRunes {
		validRunes[r] = foundRunes[r]
	}

	// Determine how many runes we can select from the top frequent runes
	availableRunes := layout.NumberOfKeys * 2
	numEssentialRunes := len(layout.EssentialRunes)
	numFrequentRunes := availableRunes - numEssentialRunes

	if numFrequentRunes < 0 {
		fmt.Println("Error: Number of essential runes exceeds available keys.")
		numFrequentRunes = 0
	}

	// Get the sorted list of runes by frequency
	sortedRunes := sortMapByValueDesc(foundRunes)

	// Add the top frequent runes to RunesToPlace, excluding essential runes already added
	i := 0
	numAdded := 0
	for numAdded < numFrequentRunes && i < len(sortedRunes) {
		r := sortedRunes[i].Key
		if _, isEssential := validRunes[r]; !isEssential {
			validRunes[r] = sortedRunes[i].Value
			numAdded++
		}
		i++
	}

	// Collect invalid runes (runes that are typeable but not in RunesToPlace)
	invalidRunes := make(map[rune]int)
	for i < len(sortedRunes) {
		r := sortedRunes[i].Key
		if _, isEssential := validRunes[r]; !isEssential {
			invalidRunes[r] = sortedRunes[i].Value
		}
		i++
	}

	for j := 0; j < len(runes); j++ {
		if isValidRune(runes[j], validRunes) {
			// Start building a quartad only if the starting rune is valid
			maxK := min(4, len(runes)-j)
			for k := 1; k <= maxK; k++ {
				// Check if all runes in the quartad are valid
				allValid := true
				for l := j; l < j+k; l++ {
					if !isValidRune(runes[l], validRunes) {
						allValid = false
						break
					}
				}
				if allValid {
					quartad := string(runes[j : j+k])
					quartads[quartad]++
				} else {
					break // No need to check longer quartads starting at j
				}
			}
		}
	}

	for r, v := range validRunes {
		runesNeedingPlacing[r] = v
	}
	for _, r := range layout.EssentialRunes {
		delete(runesNeedingPlacing, r)
	}

	orderedRunesNeedingPlacing := sortMapByValueDescToArray(runesNeedingPlacing)
	orderedInvalidRunes := sortMapByValueDescToArray(invalidRunes)

	return QuartadInfo{quartads, orderedRunesNeedingPlacing, orderedInvalidRunes}
}

func GetQuartadList(referenceTextFile string, user User) (QuartadInfo, error) {
	// Open the corpus reference file and read the entire file content
	content, err := os.ReadFile(referenceTextFile)
	if err != nil {
		return QuartadInfo{}, fmt.Errorf("error reading file: %w", err)
	}

	// Convert []byte to string and return
	text := string(content)

	// Process the text
	quartadInfo := PrepareQuartadList(text, user)
	fmt.Printf("Using %d unique runes\n", len(quartadInfo.RunesToPlace))
	fmt.Printf("%d unused runes\n", len(quartadInfo.RunesNotBeingPlaced))

	// Sort and display top 50 quartads
	sortedQuartads := sortMapByValueDesc(quartadInfo.Quartads)
	fmt.Println("Top 50 Quartads:")
	for i := 0; i < 50 && i < len(sortedQuartads); i++ {
		fmt.Printf("%q: %d\n", sortedQuartads[i].Key, sortedQuartads[i].Value)
	}

	return quartadInfo, nil
}
