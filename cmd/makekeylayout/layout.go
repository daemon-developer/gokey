package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"
)

type Layout struct {
	Name              string `json:"name"`
	SupportsOverrides bool   `json:"supports_overrides"`
	Left              Side   `json:"left"`
	Right             Side   `json:"right"`
	EssentialRunes    []rune
	FreeToPlaceRunes  int
	NumberOfKeys      int
}

type Finger int

const (
	Thumb Finger = iota
	Index
	Middle
	Ring
	Pinkie
)

type Key struct {
	MustUseUnshiftedRune rune
	MustUseShiftedRune   rune
	FreeToPlaceUnshifted bool
	FreeToPlaceShifted   bool
	AssociatedFinger     Finger
	Cost                 float64
}

type KeyPhysicalInfo struct {
	key  *Key
	side *Side
	row  int
	col  int
}

type Side struct {
	RawRows    [][]string   `json:"rows"`
	Rows       [][]Key      // Populated after processing RawRows
	ThumbHome  HomePosition `json:"thumb_home"`
	IndexHome  HomePosition `json:"index_home"`
	MiddleHome HomePosition `json:"middle_home"`
	RingHome   HomePosition `json:"ring_home"`
	PinkieHome HomePosition `json:"pinkie_home"`
}

type HomePosition struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func ReadLayout(user User) (Layout, error) {
	filename := fmt.Sprintf("keyboards/%s.json", user.Keyboard)

	var layout Layout
	data, err := os.ReadFile(filename)
	if err != nil {
		return layout, err
	}
	err = json.Unmarshal(data, &layout)
	if err != nil {
		return layout, err
	}

	// Set up essential runes and be ready to start counting the number of keys
	layout.EssentialRunes = make([]rune, 0)
	essentialRunes := make(map[rune]bool)

	// Process Rows
	layout.NumberOfKeys, layout.FreeToPlaceRunes, err = layout.Left.ProcessRows(&essentialRunes)
	if err != nil {
		return layout, err
	}
	keyCount, freeToPlaceRunes, err := layout.Right.ProcessRows(&essentialRunes)
	if err != nil {
		return layout, err
	}
	layout.NumberOfKeys += keyCount
	layout.FreeToPlaceRunes += freeToPlaceRunes

	// Process Essential Runes
	for r := range essentialRunes {
		layout.EssentialRunes = append(layout.EssentialRunes, r)
	}

	// Process the costs
	ProcessCosts(&layout.Left, user.Left)
	ProcessCosts(&layout.Right, user.Right)

	return layout, nil
}

func (s *Side) ProcessRows(essentialRunes *map[rune]bool) (int, int, error) {
	keyCount := 0
	freeToPlaceRunes := 0
	s.Rows = make([][]Key, len(s.RawRows))
	for i, row := range s.RawRows {
		keyRow := make([]Key, len(row))
		for j, keyStr := range row {
			key, err := parseKeyString(keyStr, essentialRunes)
			if err != nil {
				return 0, 0, fmt.Errorf("error parsing key at row %d, col %d: %v", i, j, err)
			}
			keyRow[j] = key
			if key.FreeToPlaceUnshifted {
				freeToPlaceRunes++
			}
			if key.FreeToPlaceShifted {
				freeToPlaceRunes++
			}
		}
		s.Rows[i] = keyRow
		keyCount += len(keyRow)
	}
	return keyCount, freeToPlaceRunes, nil
}

func parseKeyString(keyStr string, essentialRunes *map[rune]bool) (Key, error) {
	if len(keyStr) < 2 {
		return Key{}, fmt.Errorf("invalid key string: %s", keyStr)
	}

	finger, err := parseFinger(keyStr[len(keyStr)-1])
	if err != nil {
		return Key{}, err
	}

	key := Key{
		FreeToPlaceUnshifted: true,
		FreeToPlaceShifted:   true,
		AssociatedFinger:     finger,
	}

	keyContent := keyStr[:len(keyStr)-1]

	switch keyContent {
	case "*":
		// Free to place on both layers
		key.FreeToPlaceUnshifted = true
		key.FreeToPlaceShifted = true
	case "\n", "\t", "\b":
		// Control characters
		key.MustUseUnshiftedRune = runeFromString(keyContent)
		(*essentialRunes)[key.MustUseUnshiftedRune] = true
		key.FreeToPlaceUnshifted = false
		key.FreeToPlaceShifted = false
	default:
		if len(keyContent) == 1 && unicode.IsDigit(rune(keyContent[0])) {
			// Number keys
			key.MustUseUnshiftedRune = rune(keyContent[0])
			(*essentialRunes)[key.MustUseUnshiftedRune] = true
			key.FreeToPlaceUnshifted = false
			key.FreeToPlaceShifted = true
		} else if len(keyContent) == 1 && unicode.IsLetter(rune(keyContent[0])) {
			// Alphabetic letters
			lowerRune := unicode.ToLower(rune(keyContent[0]))
			upperRune := unicode.ToUpper(rune(keyContent[0]))
			key.MustUseUnshiftedRune = lowerRune
			(*essentialRunes)[key.MustUseUnshiftedRune] = true
			key.MustUseShiftedRune = upperRune
			(*essentialRunes)[key.MustUseShiftedRune] = true

			key.FreeToPlaceUnshifted = false
			key.FreeToPlaceShifted = false
		} else {
			// Other symbols or fixed assignments
			key.MustUseUnshiftedRune = runeFromString(keyContent)
			key.FreeToPlaceUnshifted = false
			key.FreeToPlaceShifted = true // Assuming shifted layer is free
		}
	}

	return key, nil
}

func runeFromString(s string) rune {
	// Handle escaped characters like "\n", "\t", "\b"
	switch s {
	case "\\n":
		return '\n'
	case "\\t":
		return '\t'
	case "\\b":
		return '\b'
	default:
		return rune(s[0])
	}
}

func parseFinger(fingerChar byte) (Finger, error) {
	switch fingerChar {
	case 'T':
		return Thumb, nil
	case 'I':
		return Index, nil
	case 'M':
		return Middle, nil
	case 'R':
		return Ring, nil
	case 'P':
		return Pinkie, nil
	default:
		return -1, fmt.Errorf("invalid finger character: %c", fingerChar)
	}
}

func ProcessCosts(side *Side, hand Hand) {
	for r := 0; r < len(side.Rows); r++ {
		for c := 0; c < len(side.Rows[r]); c++ {
			side.Rows[r][c].Cost = calculateFingerCost(r, c, hand, *side)
		}
	}
}

func calculateFingerCost(row, col int, hand Hand, side Side) float64 {
	key := side.Rows[row][col]
	homePosition := getFingerHomePosition(key.AssociatedFinger, side)
	deltaRow := row - homePosition.Row
	deltaCol := col - homePosition.Col
	fingerCosts := getFingerCost(key.AssociatedFinger, hand)
	baseCost := fingerCosts.Cost
	baseCost += math.Abs(float64(deltaCol)) * fingerCosts.HCost
	if deltaRow < 0 {
		baseCost += math.Abs(float64(deltaRow)) * fingerCosts.UpCost
	} else if deltaRow > 0 {
		baseCost += math.Abs(float64(deltaRow)) * fingerCosts.DownCost
	}
	return baseCost
}

func getFingerHomePosition(finger Finger, side Side) HomePosition {
	switch finger {
	case Thumb:
		return side.ThumbHome
	case Index:
		return side.IndexHome
	case Middle:
		return side.MiddleHome
	case Ring:
		return side.RingHome
	case Pinkie:
		return side.PinkieHome
	default:
		return HomePosition{}
	}
}

func getFingerCost(finger Finger, hand Hand) FingerCost {
	switch finger {
	case Thumb:
		return hand.Thumb
	case Index:
		return hand.Index
	case Middle:
		return hand.Middle
	case Ring:
		return hand.Ring
	case Pinkie:
		return hand.Pinkie
	default:
		return FingerCost{}
	}
}

func (layout *Layout) mapRunesToPhysicalKeyInfo() map[rune]*KeyPhysicalInfo {
	keyMap := make(map[rune]*KeyPhysicalInfo)

	for r, _ := range layout.Left.Rows {
		for c, key := range layout.Left.Rows[r] {
			if !key.FreeToPlaceUnshifted {
				keyMap[key.MustUseUnshiftedRune] = &KeyPhysicalInfo{key: &layout.Left.Rows[r][c], side: &layout.Left, row: r, col: c}
			}
			if !key.FreeToPlaceShifted {
				keyMap[key.MustUseShiftedRune] = &KeyPhysicalInfo{key: &layout.Left.Rows[r][c], side: &layout.Left, row: r, col: c}
			}
		}
	}

	for r, _ := range layout.Right.Rows {
		for c, key := range layout.Right.Rows[r] {
			if !key.FreeToPlaceUnshifted {
				keyMap[key.MustUseUnshiftedRune] = &KeyPhysicalInfo{key: &layout.Right.Rows[r][c], side: &layout.Right, row: r, col: c}
			}
			if !key.FreeToPlaceShifted {
				keyMap[key.MustUseShiftedRune] = &KeyPhysicalInfo{key: &layout.Right.Rows[r][c], side: &layout.Right, row: r, col: c}
			}
		}
	}

	return keyMap
}

// DeepCopyLayout creates an efficient deep copy of the given Layout
func DeepCopyLayout(layout Layout) Layout {
	copiedLayout := Layout{
		Name:              layout.Name,
		SupportsOverrides: layout.SupportsOverrides,
		FreeToPlaceRunes:  layout.FreeToPlaceRunes,
		NumberOfKeys:      layout.NumberOfKeys,
	}

	// Deep copy EssentialRunes
	copiedLayout.EssentialRunes = make([]rune, len(layout.EssentialRunes))
	copy(copiedLayout.EssentialRunes, layout.EssentialRunes)

	// Deep copy Left and Right sides
	copiedLayout.Left = deepCopySide(layout.Left)
	copiedLayout.Right = deepCopySide(layout.Right)

	return copiedLayout
}

func deepCopySide(side Side) Side {
	copiedSide := Side{
		RawRows:    make([][]string, len(side.RawRows)),
		Rows:       make([][]Key, len(side.Rows)),
		ThumbHome:  side.ThumbHome,
		IndexHome:  side.IndexHome,
		MiddleHome: side.MiddleHome,
		RingHome:   side.RingHome,
		PinkieHome: side.PinkieHome,
	}

	// Deep copy RawRows
	for i, row := range side.RawRows {
		copiedSide.RawRows[i] = make([]string, len(row))
		copy(copiedSide.RawRows[i], row)
	}

	// Deep copy Rows
	for i, row := range side.Rows {
		copiedSide.Rows[i] = make([]Key, len(row))
		copy(copiedSide.Rows[i], row)
	}

	return copiedSide
}

func (layout *Layout) String() string {
	return layout.stringInternal(false)
}

func (layout *Layout) StringWithCosts() string {
	return layout.stringInternal(true)
}

func (layout *Layout) stringInternal(costs bool) string {
	var sb strings.Builder

	// Write the layout name
	sb.WriteString(fmt.Sprintf("Layout: %s\n\n", layout.Name))

	// Determine the maximum number of rows
	maxRows := len(layout.Left.Rows)
	if len(layout.Right.Rows) > maxRows {
		maxRows = len(layout.Right.Rows)
	}

	// Generate the layout visualization
	for i := 0; i < maxRows; i++ {
		leftRow := ""
		rightRow := ""

		if i < len(layout.Left.Rows) {
			leftRow += visualizeRow(layout.Left.Rows[i], costs)
		}
		if i < len(layout.Right.Rows) {
			rightRow += visualizeRow(layout.Right.Rows[i], costs)
		}

		if !costs {
			sb.WriteString(fmt.Sprintf("%25s  |  %s\n", leftRow, rightRow))
		} else {
			sb.WriteString(fmt.Sprintf("%42s  |  %s\n", leftRow, rightRow))
		}
	}

	return sb.String()
}

func visualizeRow(row []Key, costs bool) string {
	var keys []string
	var specialCharMap = map[rune]rune{
		'\t': '⇥', // Tab
		'\b': '⌫', // Backspace
		'\r': '↵', // Return
		'\n': '↵', // Newline (often the same as Return)
		' ':  '␣', // Space
		// Add more special characters as needed
	}
	unassignedKeySymbol := rune('☒')

	if !costs {
		for _, key := range row {
			if key.MustUseUnshiftedRune != 0 {
				displayRune := key.MustUseUnshiftedRune
				if mappedRune, ok := specialCharMap[displayRune]; ok {
					displayRune = mappedRune
				}
				keys = append(keys, fmt.Sprintf("[%c]", displayRune))
			} else {
				keys = append(keys, fmt.Sprintf("[%c]", unassignedKeySymbol))
			}
		}
	} else {
		for _, key := range row {
			keys = append(keys, fmt.Sprintf("[%1.2f]", key.Cost))
		}
	}

	return strings.Join(keys, " ")
}
