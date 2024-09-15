package main

import (
	"fmt"
	"os"
	"unicode"
)

type QuartadList map[string]int

type QuartadInfo struct {
	Quartads        QuartadList
	RunesOnKeyboard []rune
}

func isTypeableRune(r rune) bool {
	// Check for common typeable characters
	if (unicode.IsPrint(r) || r == '\t' || r == '\n') && r < 128 {
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
	quartads := make(QuartadList)
	runes := []rune(s)
	n := len(runes)

	// Ensure essential runes are included
	for _, r := range layout.EssentialRunes {
		if unicode.IsLetter(r) {
			foundRunes[unicode.ToUpper(r)] = 0
			foundRunes[unicode.ToLower(r)] = 0
		} else {
			foundRunes[r] = 0
		}
	}

	// The essential from the layout are essential meet the needs of hardcoded
	// keys on that layout. We now also need to add any the user has asked for
	reqRunes := []rune(user.Required)
	for _, r := range reqRunes {
		if unicode.IsLetter(r) {
			foundRunes[unicode.ToUpper(r)] = 0
			foundRunes[unicode.ToLower(r)] = 0
		} else {
			foundRunes[r] = 0
		}

	}

	// Count the frequency of all the runes
	for i := 0; i < n; i++ {
		if isTypeableRune(runes[i]) {
			if _, ok := foundRunes[runes[i]]; !ok {
				fmt.Printf("Found rune '%c'\n", RuneDisplayVersion(runes[i]))
			}
			foundRunes[runes[i]]++
		}
	}

	// Map runes onto the keyboard in usage order (with essential first) so
	// we can build quartads with what we know are on the keyboard
	validRunes := layout.AssignRunesToKeys(foundRunes, user)

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

	runesOnKeyboard := make([]rune, len(validRunes))
	i := 0
	for _, r := range validRunes {
		runesOnKeyboard[i] = rune(r)
		i++
	}

	return QuartadInfo{quartads, runesOnKeyboard}
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
	fmt.Printf("Using %d unique runes\n", len(quartadInfo.RunesOnKeyboard))

	// Sort and display top 50 quartads
	sortedQuartads := sortMapByValueDesc(quartadInfo.Quartads)
	fmt.Println("Top 50 Quartads:")
	for i := 0; i < 50 && i < len(sortedQuartads); i++ {
		fmt.Printf("%q: %d\n", sortedQuartads[i].Key, sortedQuartads[i].Value)
	}

	return quartadInfo, nil
}
