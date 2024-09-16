package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Locale struct {
	unshiftedToShifted map[rune]rune
	shiftedToUnshifted map[rune]rune
}

func LoadUserLocale(locateFile string) (Locale, error) {
	filename := fmt.Sprintf("locale/%s.json", locateFile)

	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return Locale{}, fmt.Errorf("error reading file: %w", err)
	}

	// Parse JSON into a temporary map[string]string
	var tempMap map[string]string
	err = json.Unmarshal(data, &tempMap)
	if err != nil {
		return Locale{}, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Convert the temporary map to map[rune]rune
	unshiftedToShifted := make(map[rune]rune)
	shiftedToUnshifted := make(map[rune]rune)
	for k, v := range tempMap {
		kRunes := []rune(k)
		vRunes := []rune(v)
		if len(kRunes) == 1 && len(vRunes) == 1 {
			unshiftedToShifted[kRunes[0]] = vRunes[0]
			shiftedToUnshifted[vRunes[0]] = kRunes[0]
		} else {
			return Locale{}, fmt.Errorf("invalid key-value pair: %s: %s (must be single Unicode characters)", k, v)
		}
	}

	return Locale{unshiftedToShifted: unshiftedToShifted, shiftedToUnshifted: shiftedToUnshifted}, nil
}
