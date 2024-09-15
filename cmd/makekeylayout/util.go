package main

import (
	"cmp"
	"sort"
)

type KeyValue[K cmp.Ordered, V cmp.Ordered] struct {
	Key   K
	Value V
}

func sortMapByValueDesc[K cmp.Ordered, V cmp.Ordered](m map[K]V) []KeyValue[K, V] {
	sorted := make([]KeyValue[K, V], 0, len(m))
	for k, v := range m {
		sorted = append(sorted, KeyValue[K, V]{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})
	return sorted
}

func sortMapByValueDescToArray[K cmp.Ordered, V cmp.Ordered](m map[K]V) []K {
	sorted := make([]KeyValue[K, V], 0, len(m))
	for k, v := range m {
		sorted = append(sorted, KeyValue[K, V]{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	var keys []K
	for _, kv := range sorted {
		keys = append(keys, kv.Key)
	}
	return keys
}

func RuneDisplayVersion(r rune) rune {
	var specialCharMap = map[rune]rune{
		'\t':    '⇥', // Tab
		'\b':    '⌫', // Backspace
		'\r':    '↵', // Return
		'\n':    '↵', // Newline (often the same as Return)
		' ':     '␣', // Space
		rune(0): '☒', // Unassigned
		// Add more special characters as needed
	}

	if mappedRune, ok := specialCharMap[r]; ok {
		return mappedRune
	}

	return r
}
