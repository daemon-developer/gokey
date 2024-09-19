package main

import (
	"cmp"
)

type KeyValue[K cmp.Ordered, V cmp.Ordered] struct {
	Key   K
	Value V
}

func randomizeMapToArray[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Randomize the slice by shuffling it
	return randomizeSlice(keys)
}

func randomizeSlice[T any](slice []T) []T {
	for i := len(slice) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

func RuneDisplayVersion(r rune) rune {
	var specialCharMap = map[rune]rune{
		'\t':                '⇥', // Tab
		'\b':                '⌫', // Backspace
		'\r':                '↵', // Return
		'\n':                '↵', // Newline (often the same as Return)
		' ':                 '␣', // Space
		rune(ShiftModifier): '⇧', // SHIFT symbol
		rune(CtrlModifier):  '^', // CTRL symbol
		rune(AltModifier):   '⌥', // ALT symbol
		rune(0):             '⋀', // Unassigned
		// Add more special characters as needed
	}

	if mappedRune, ok := specialCharMap[r]; ok {
		return mappedRune
	}

	return r
}

func AbsI(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}
