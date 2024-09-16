package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
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
	UnshiftedRune   rune
	ShiftedRune     rune
	UnshiftedIsFree bool
	ShiftedIsFree   bool
}

type KeyPhysicalInfo struct {
	key              *Key
	swappable        bool
	hand             *Side
	associatedFinger Finger
	cost             float64
	row              int
	col              int
	horzDeltaToHome  int
	vertDeltaToHome  int
}

type Side struct {
	RawRows    [][]string          `json:"rows"`
	Rows       [][]KeyPhysicalInfo // Populated after processing RawRows
	ThumbHome  HomePosition        `json:"thumb_home"`
	IndexHome  HomePosition        `json:"index_home"`
	MiddleHome HomePosition        `json:"middle_home"`
	RingHome   HomePosition        `json:"ring_home"`
	PinkieHome HomePosition        `json:"pinkie_home"`
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
	layout.NumberOfKeys, layout.FreeToPlaceRunes, err = layout.ProcessRows(&layout.Left, &essentialRunes, layout.SupportsOverrides, user.Locale)
	if err != nil {
		return layout, err
	}
	keyCount, freeToPlaceRunes, err := layout.ProcessRows(&layout.Right, &essentialRunes, layout.SupportsOverrides, user.Locale)
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
	layout.ProcessCosts(&layout.Left, user.Left)
	layout.ProcessCosts(&layout.Right, user.Right)

	return layout, nil
}

func (layout *Layout) ProcessRows(s *Side, essentialRunes *map[rune]bool, supportOverrides bool, locale Locale) (int, int, error) {
	keyCount := 0
	freeToPlaceRunes := 0
	s.Rows = make([][]KeyPhysicalInfo, len(s.RawRows))
	for r, row := range s.RawRows {
		keyRow := make([]KeyPhysicalInfo, len(row))
		for c, keyStr := range row {
			keyInfo, err := layout.parseKeyString(s, r, c, keyStr, essentialRunes, supportOverrides, locale)
			if err != nil {
				return 0, 0, fmt.Errorf("error parsing keyInfo at row %d, col %d: %v", r, c, err)
			}
			keyRow[c] = keyInfo
			if keyInfo.key.UnshiftedIsFree {
				freeToPlaceRunes++
			}
			if keyInfo.key.ShiftedIsFree {
				freeToPlaceRunes++
			}
		}
		s.Rows[r] = keyRow
		keyCount += len(keyRow)
	}
	return keyCount, freeToPlaceRunes, nil
}

func (layout *Layout) parseKeyString(s *Side, r, c int, keyStr string, essentialRunes *map[rune]bool, supportOverrides bool, locale Locale) (KeyPhysicalInfo, error) {
	if len(keyStr) < 2 {
		return KeyPhysicalInfo{}, fmt.Errorf("invalid key string: %s", keyStr)
	}

	finger, err := parseFinger(keyStr[len(keyStr)-1])
	if err != nil {
		return KeyPhysicalInfo{}, err
	}

	key := Key{
		UnshiftedIsFree: true,
		ShiftedIsFree:   true,
	}

	keyInfo := KeyPhysicalInfo{
		key:              &key,
		swappable:        false,
		hand:             s,
		associatedFinger: finger,
		cost:             0,
		row:              r,
		col:              c,
		horzDeltaToHome:  0,
		vertDeltaToHome:  0,
	}

	keyContent := keyStr[:len(keyStr)-1]

	switch keyContent {
	case "*":
		// Free to place on both layers
		key.UnshiftedIsFree = true
		key.ShiftedIsFree = true
		keyInfo.swappable = true
	case "\n", "\t", "\b", " ":
		// Control characters
		key.UnshiftedRune = runeFromString(keyContent)
		key.ShiftedRune = key.UnshiftedRune
		(*essentialRunes)[key.UnshiftedRune] = true
		key.UnshiftedIsFree = false
		key.ShiftedIsFree = false
	default:
		if len(keyContent) == 1 && unicode.IsDigit(rune(keyContent[0])) {
			// Number keys
			key.UnshiftedRune = rune(keyContent[0])
			(*essentialRunes)[key.UnshiftedRune] = true
			key.UnshiftedIsFree = false
			key.ShiftedIsFree = true
			if !supportOverrides {
				shiftedRune := locale.unshiftedToShifted[key.UnshiftedRune]
				key.ShiftedRune = shiftedRune
				(*essentialRunes)[key.ShiftedRune] = true
				key.ShiftedIsFree = false
			}
		} else if len(keyContent) == 1 && unicode.IsLetter(rune(keyContent[0])) {
			// Alphabetic letters
			lowerRune := unicode.ToLower(rune(keyContent[0]))
			upperRune := unicode.ToUpper(rune(keyContent[0]))
			key.UnshiftedRune = lowerRune
			(*essentialRunes)[key.UnshiftedRune] = true
			key.ShiftedRune = upperRune
			(*essentialRunes)[key.ShiftedRune] = true
			key.UnshiftedIsFree = false
			key.ShiftedIsFree = false
		} else {
			// Other symbols or fixed assignments
			key.UnshiftedRune = runeFromString(keyContent)
			key.UnshiftedIsFree = false
			key.ShiftedIsFree = true // Assuming shifted layer is free
		}
	}

	return keyInfo, nil
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

func (layout *Layout) ProcessCosts(side *Side, hand Hand) {
	for r := 0; r < len(side.Rows); r++ {
		for c := 0; c < len(side.Rows[r]); c++ {
			keyInfo := &side.Rows[r][c]
			keyInfo.cost, keyInfo.vertDeltaToHome, keyInfo.horzDeltaToHome = calculateFingerCost(r, c, hand, *side)
		}
	}
}

func calculateFingerCost(row, col int, hand Hand, side Side) (float64, int, int) {
	key := side.Rows[row][col]
	homePosition := getFingerHomePosition(key.associatedFinger, side)
	deltaRow := row - homePosition.Row
	deltaCol := col - homePosition.Col
	fingerCosts := getFingerCost(key.associatedFinger, hand)
	baseCost := fingerCosts.Cost
	baseCost += math.Abs(float64(deltaCol)) * fingerCosts.HCost
	if deltaRow < 0 {
		baseCost += math.Abs(float64(deltaRow)) * fingerCosts.UpCost
	} else if deltaRow > 0 {
		baseCost += math.Abs(float64(deltaRow)) * fingerCosts.DownCost
	}
	return baseCost, deltaRow, deltaCol
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

	for r := range layout.Left.Rows {
		for c, keyInfo := range layout.Left.Rows[r] {
			if !keyInfo.key.UnshiftedIsFree {
				keyMap[keyInfo.key.UnshiftedRune] = &layout.Left.Rows[r][c]
			}
			if !keyInfo.key.ShiftedIsFree {
				keyMap[keyInfo.key.ShiftedRune] = &layout.Left.Rows[r][c]
			}
		}
	}

	for r := range layout.Right.Rows {
		for c, keyInfo := range layout.Right.Rows[r] {
			if !keyInfo.key.UnshiftedIsFree {
				keyMap[keyInfo.key.UnshiftedRune] = &layout.Right.Rows[r][c]
			}
			if !keyInfo.key.ShiftedIsFree {
				keyMap[keyInfo.key.ShiftedRune] = &layout.Right.Rows[r][c]
			}
		}
	}

	return keyMap
}

func (layout *Layout) GetSwappableKeys() []*Key {
	var keys []*Key
	for r := range layout.Left.Rows {
		for _, keyInfo := range layout.Left.Rows[r] {
			if keyInfo.swappable {
				keys = append(keys, keyInfo.key)
			}
		}
	}
	for r := range layout.Right.Rows {
		for _, keyInfo := range layout.Right.Rows[r] {
			if keyInfo.swappable {
				keys = append(keys, keyInfo.key)
			}
		}
	}
	return keys
}

func (layout *Layout) AssignRunesToKeys(foundRunes map[rune]int, user User) (map[rune]int, map[rune]int) {
	orderedFoundRunes := sortMapByValueDescToArray(foundRunes)
	for _, r := range orderedFoundRunes {
		fmt.Printf("Rune '%c' used %d times\n", RuneDisplayVersion(r), foundRunes[r])
	}
	orderedKeyInfos := layout.getOrderedKeysByCost()

	// Before we start assigning keys, we'll run over the current layout and mark that certain runes are
	// already assigned
	assignedRunes := make(map[rune]int)
	assignedShiftedRunes := make(map[rune]int)
	alreadyAssigned := make(map[rune]bool)
	for _, keyInfo := range orderedKeyInfos {
		if keyInfo.key.UnshiftedIsFree == false {
			r := keyInfo.key.UnshiftedRune
			alreadyAssigned[r] = true
			assignedRunes[r] = foundRunes[r]
		}
		if keyInfo.key.ShiftedIsFree == false {
			r := keyInfo.key.ShiftedRune
			alreadyAssigned[r] = true
			assignedRunes[r] = foundRunes[r]
		}
	}

	i := 0
	k := 0
	for i < len(orderedFoundRunes) && k < len(orderedKeyInfos) {
		keyInfo := orderedKeyInfos[k]
		if i >= len(orderedFoundRunes) {
			fmt.Println("Breaking loop as out of ordered runes")
			break
		}

		// Get the rune to assign
		runeToAssign := orderedFoundRunes[i]

		// Check if it's already assigned
		if alreadyAssigned[runeToAssign] {
			i++
			continue
		}

		if unicode.IsLetter(runeToAssign) {
			if keyInfo.key.UnshiftedIsFree == false {
				k++
				fmt.Printf("Skipping key %d,%d which has rune '%c' already\n", keyInfo.row, keyInfo.col, RuneDisplayVersion(keyInfo.key.UnshiftedRune))
				continue
			}
			lowerAlpha := unicode.ToLower(runeToAssign)
			if alreadyAssigned[lowerAlpha] == false {
				// If it's a letter, assign lower and upper case
				upperAlpha := unicode.ToUpper(runeToAssign)
				fmt.Printf("Handling '%c' & '%c' as letter\n", upperAlpha, lowerAlpha)
				keyInfo.key.UnshiftedRune = lowerAlpha
				keyInfo.key.ShiftedRune = upperAlpha
				keyInfo.key.UnshiftedIsFree = false
				keyInfo.key.ShiftedIsFree = false
				alreadyAssigned[lowerAlpha] = true
				assignedRunes[lowerAlpha] = foundRunes[lowerAlpha]
				assignedRunes[upperAlpha] = foundRunes[upperAlpha]
				assignedShiftedRunes[upperAlpha] = foundRunes[upperAlpha]
			}
		} else {
			// Handle symbol assignment
			if layout.SupportsOverrides {
				if keyInfo.key.UnshiftedIsFree {
					keyInfo.key.UnshiftedRune = runeToAssign
					keyInfo.key.UnshiftedIsFree = false
					alreadyAssigned[runeToAssign] = true
					assignedRunes[runeToAssign] = foundRunes[runeToAssign]
				} else {
					keyInfo.key.ShiftedRune = runeToAssign
					keyInfo.key.ShiftedIsFree = false
					alreadyAssigned[runeToAssign] = true
					assignedRunes[runeToAssign] = foundRunes[runeToAssign]
					assignedShiftedRunes[runeToAssign] = foundRunes[runeToAssign]
				}
			} else {
				if keyInfo.key.UnshiftedIsFree == false {
					k++
					fmt.Printf("Skipping key %d,%d which has rune '%c' already\n", keyInfo.row, keyInfo.col, RuneDisplayVersion(keyInfo.key.UnshiftedRune))
					continue
				}
				// Use locale-based symbol mapping if overrides aren't supported
				if shiftedRune, exists := user.Locale.unshiftedToShifted[runeToAssign]; exists {
					fmt.Printf("Handling '%c' & '%c' as unshifted symbol\n", runeToAssign, shiftedRune)
					keyInfo.key.UnshiftedRune = runeToAssign
					keyInfo.key.ShiftedRune = shiftedRune
					keyInfo.key.UnshiftedIsFree = false
					keyInfo.key.ShiftedIsFree = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[shiftedRune] = true
					assignedRunes[runeToAssign] = foundRunes[runeToAssign]
					assignedRunes[shiftedRune] = foundRunes[shiftedRune]
					assignedShiftedRunes[shiftedRune] = foundRunes[shiftedRune]
				} else if unshiftedRune, exists := user.Locale.shiftedToUnshifted[runeToAssign]; exists {
					fmt.Printf("Handling '%c' & '%c' as shifted symbol\n", runeToAssign, unshiftedRune)
					keyInfo.key.UnshiftedRune = unshiftedRune
					keyInfo.key.ShiftedRune = runeToAssign
					keyInfo.key.UnshiftedIsFree = false
					keyInfo.key.ShiftedIsFree = false
					alreadyAssigned[runeToAssign] = true
					alreadyAssigned[unshiftedRune] = true
					assignedRunes[unshiftedRune] = foundRunes[unshiftedRune]
					assignedRunes[runeToAssign] = foundRunes[runeToAssign]
					assignedShiftedRunes[runeToAssign] = foundRunes[runeToAssign]
				} else {
					// Not actually symbols so will be things like ENTER or backspace
					fmt.Printf("Handling '%c' as non-symbol\n", RuneDisplayVersion(runeToAssign))
					keyInfo.key.UnshiftedRune = runeToAssign
					keyInfo.key.ShiftedRune = runeToAssign
					keyInfo.key.UnshiftedIsFree = false
					keyInfo.key.ShiftedIsFree = false
					alreadyAssigned[runeToAssign] = true
					assignedRunes[runeToAssign] = foundRunes[runeToAssign]
				}
			}
		}

		i++
		if keyInfo.key.UnshiftedIsFree == false {
			k++
		}
	}

	for r, v := range assignedRunes {
		fmt.Printf("Assigned rune '%c' to a key (rune was used %d times)\n", RuneDisplayVersion(r), v)
	}

	fmt.Println(layout.String())

	// Return the runes we have assigned to keys
	return assignedRunes, assignedShiftedRunes
}

func (layout *Layout) getOrderedKeysByCost() []*KeyPhysicalInfo {
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

func (layout *Layout) Duplicate() Layout {
	return layout.deepCopy()
}

func (layout *Layout) deepCopy() Layout {
	// Create a new Layout object
	copyLayout := *layout

	// Deep copy the left and right sides
	copyLayout.Left = layout.Left.DeepCopy()
	copyLayout.Right = layout.Right.DeepCopy()

	// Deep copy the EssentialRunes slice
	copyLayout.EssentialRunes = make([]rune, len(layout.EssentialRunes))
	copy(copyLayout.EssentialRunes, layout.EssentialRunes)

	return copyLayout
}

func (s *Side) DeepCopy() Side {
	// Create a new Side object
	copySide := *s

	// Deep copy the RawRows slice of slices
	copySide.RawRows = make([][]string, len(s.RawRows))
	for i := range s.RawRows {
		copySide.RawRows[i] = make([]string, len(s.RawRows[i]))
		copy(copySide.RawRows[i], s.RawRows[i])
	}

	// Deep copy the Rows slice of slices
	copySide.Rows = make([][]KeyPhysicalInfo, len(s.Rows))
	for i := range s.Rows {
		copySide.Rows[i] = make([]KeyPhysicalInfo, len(s.Rows[i]))
		for j := range s.Rows[i] {
			copySide.Rows[i][j] = s.Rows[i][j].DeepCopy()
		}
	}

	return copySide
}

func (kpi *KeyPhysicalInfo) DeepCopy() KeyPhysicalInfo {
	copyKpi := *kpi

	// Deep copy the key pointer
	if kpi.key != nil {
		newKey := *kpi.key
		copyKpi.key = &newKey
	}

	// No need to copy primitives like associatedFinger, cost, etc.
	return copyKpi
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

func visualizeRow(row []KeyPhysicalInfo, costs bool) string {
	var keys []string

	if !costs {
		for _, keyInfo := range row {
			if keyInfo.key.UnshiftedRune != 0 {
				displayRune := RuneDisplayVersion(unicode.ToUpper(keyInfo.key.UnshiftedRune))
				keys = append(keys, fmt.Sprintf("[%c]", displayRune))
			} else {
				keys = append(keys, fmt.Sprintf("[%c]", ' '))
			}
		}
	} else {
		for _, keyInfo := range row {
			keys = append(keys, fmt.Sprintf("[%1.2f]", keyInfo.cost))
		}
	}

	return strings.Join(keys, " ")
}
