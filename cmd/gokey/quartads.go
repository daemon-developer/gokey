package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Modifier rune

const (
	NoModifier    Modifier = 0xe00
	ShiftModifier Modifier = 0xE001 // Define the SHIFT modifier as a rune
	CtrlModifier  Modifier = 0xE002 // Example: Define the CTRL modifier as another rune
	AltModifier   Modifier = 0xE003 // Example: Define the ALT modifier as another rune
)

// Quartad struct represents a sequence of up to 4 runes (key presses)
// along with their associated modifiers
type Quartad struct {
	length    int         // Number of actual runes in the quartad (1 to 4)
	runes     [4]rune     // Up to 4 runes (key presses)
	modifiers [4]Modifier // Associated modifiers for each rune (e.g., Shift, Ctrl)
}

type QuartadList map[Quartad]int

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
	reqRunes := user.Required
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
	runesOnKeyboard, shiftedRunesOnKeyboard := layout.AssignRunesToKeys(foundRunes, user)

	for j := 0; j < len(runes); j++ {
		if isValidRune(runes[j], runesOnKeyboard) {
			// Start building a quartad only if the starting rune is valid
			maxK := min(4, len(runes)-j)
			for k := 1; k <= maxK; k++ {
				// Check if all runes in the quartad are valid
				allValid := true
				for l := j; l < j+k; l++ {
					if !isValidRune(runes[l], runesOnKeyboard) {
						allValid = false
						break
					}
				}
				if allValid {
					quartad := MakeQuartad(string(runes[j:j+k]), shiftedRunesOnKeyboard)
					quartads[quartad]++
				} else {
					break // No need to check longer quartads starting at j
				}
			}
		}
	}

	runesOnKeyboardResult := make([]rune, len(runesOnKeyboard))
	i := 0
	for _, r := range runesOnKeyboard {
		runesOnKeyboardResult[i] = rune(r)
		i++
	}

	return QuartadInfo{quartads, runesOnKeyboardResult}
}

func GetQuartadList(referenceTextFiles []string, user User) (QuartadInfo, error) {
	// Read all the requested corpus files and concatenate them
	sb := strings.Builder{}

	for _, referenceTextFile := range referenceTextFiles {
		// Open the corpus reference file and read the entire file content
		content, err := os.ReadFile(referenceTextFile)
		if err != nil {
			return QuartadInfo{}, fmt.Errorf("error reading file: %w", err)
		}

		// Convert []byte to string and return
		text := string(content)
		sb.WriteString(text)
	}

	// Process the text
	quartadInfo := PrepareQuartadList(sb.String(), user)
	fmt.Printf("Using %d unique runes\n", len(quartadInfo.RunesOnKeyboard))

	// Sort and display top 50 quartads
	sortedQuartads := SortQuartadsMapByKeyDesc(quartadInfo.Quartads)
	fmt.Println("Top 50 Quartads:")
	for i := 0; i < 50 && i < len(sortedQuartads); i++ {
		fmt.Printf("%q: %d\n", sortedQuartads[i].Key.String(), sortedQuartads[i].Value)
	}

	return quartadInfo, nil
}

// Equals checks if two Quartads are equal, considering the length of the quartad.
// It compares both the runes and modifiers for each valid element.
func (q Quartad) Equals(other Quartad) bool {
	// If the lengths are different, the quartads can't be equal
	if q.length != other.length {
		return false
	}

	// Compare each rune and modifier in the quartad up to the specified length
	for i := 0; i < q.length; i++ {
		if q.runes[i] != other.runes[i] || q.modifiers[i] != other.modifiers[i] {
			return false
		}
	}
	return true
}

// NotEquals checks if two Quartads are not equal.
func (q Quartad) NotEquals(other Quartad) bool {
	return !q.Equals(other)
}

func (q Quartad) Cmp(other Quartad) int {
	// First, compare the lengths
	if q.length < other.length {
		return -1
	}
	if q.length > other.length {
		return 1
	}

	// Compare the runes lexicographically, up to the specified length
	for i := 0; i < q.length; i++ {
		if q.runes[i] < other.runes[i] {
			return -1
		}
		if q.runes[i] > other.runes[i] {
			return 1
		}
	}

	// If runes are the same, compare the modifiers lexicographically
	for i := 0; i < q.length; i++ {
		if q.modifiers[i] < other.modifiers[i] {
			return -1
		}
		if q.modifiers[i] > other.modifiers[i] {
			return 1
		}
	}

	// If length, runes, and modifiers are all the same, return 0 (equal)
	return 0
}
func SortQuartadsMapByKeyDesc(m map[Quartad]int) []struct {
	Key   Quartad
	Value int
} {
	sorted := make([]struct {
		Key   Quartad
		Value int
	}, 0, len(m))

	for k, v := range m {
		sorted = append(sorted, struct {
			Key   Quartad
			Value int
		}{k, v})
	}

	// Sort by value in descending order
	sort.Slice(sorted, func(i, j int) bool {
		// Sort by value in descending order
		if sorted[i].Value != sorted[j].Value {
			return sorted[i].Value > sorted[j].Value
		}
		// If values are equal, use Quartad comparison as a tiebreaker
		return sorted[i].Key.Cmp(sorted[j].Key) > 0
	})

	return sorted
}

func (q Quartad) Len() int {
	return q.length
}

func addRuneToQuartad(q *Quartad, r rune, m Modifier) {
	if q.length < len(q.runes) {
		q.runes[q.length] = r
		q.modifiers[q.length] = m
		q.length++
	}
}

func (q Quartad) String() string {
	sb := strings.Builder{}
	for i := range q.length {
		sb.WriteRune(RuneDisplayVersion(q.runes[i]))
	}
	return sb.String()
}

func (q Quartad) GetRune(i int) rune {
	return q.runes[i]
}

func MakeQuartad(s string, shiftedRunes map[rune]int) Quartad {
	var q Quartad
	for _, r := range s {
		modifier := NoModifier
		if _, isShifted := shiftedRunes[r]; isShifted {
			modifier = ShiftModifier
		}
		addRuneToQuartad(&q, r, modifier)
	}
	return q
}
