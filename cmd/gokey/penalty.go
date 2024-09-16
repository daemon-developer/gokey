package main

// KeyPenalty defines a penalty rule.
type KeyPenalty struct {
	Name     string
	Function PenaltyFunc
	Cost     float64
}

type PenaltyFunc func(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64

type KeyPenaltyResult struct {
	Name     string
	Total    float64
	HighKeys map[Quartad]float64
	Info     *KeyPenalty
}

// InitPenaltyRules initializes the penalty rules.
func InitPenaltyRules(user User) []KeyPenalty {
	return []KeyPenalty{
		{Name: "Base", Function: calcBasePenalty, Cost: 1.0},
		{Name: "SFB", Function: calcSFBPenalty, Cost: user.Penalties.SFB},
		{Name: "Vertical finger travel", Function: calcVerticalFingerTravelPenalty, Cost: user.Penalties.VerticalFingerTravel},
		{Name: "Long SFB", Function: calcLongSFBPenalty, Cost: user.Penalties.LongSFB},
		{Name: "Lateral Stretch", Function: calcLateralStretchPenalty, Cost: user.Penalties.LateralStretch},
		{Name: "Pinky/Ring Stretch", Function: calcPinkyRingStretchPenalty, Cost: user.Penalties.PinkyRingStretch},
		{Name: "Roll reversal", Function: calcRollReversalPenalty, Cost: user.Penalties.RollReversal},
		{Name: "Hand repetition", Function: calcHandRepetitionPenalty, Cost: user.Penalties.HandRepetition},
		{Name: "Hand alternation", Function: calcHandAlternationPenalty, Cost: user.Penalties.HandAlternation},
		{Name: "Outward roll", Function: calcOutwardRollPenalty, Cost: user.Penalties.OutwardRoll},
		{Name: "Inward roll", Function: calcInwardRollPenalty, Cost: user.Penalties.InwardRoll},
		{Name: "Scissor motion", Function: calcScissorMotionPenalty, Cost: user.Penalties.ScissorMotion},
		{Name: "Row change in roll", Function: calcRowChangeInRollPenalty, Cost: user.Penalties.RowChangeInRoll},
		{Name: "Base modifier", Function: calcBaseModifierPenalty, Cost: 1.0},
		{Name: "Same finger modifier", Function: calcSameFingerModifierPenalty, Cost: user.Penalties.SameFingerModifier},
	}
}

func calcBasePenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil {
		return 0.0
	}
	return curr.cost * cost
}

func calcSFBPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand && curr.associatedFinger == old1.associatedFinger {
		if curr.key != old1.key {
			return cost
		}
	}
	return 0.0
}

func calcVerticalFingerTravelPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand {
		delta := AbsI(curr.row - old1.row)
		if delta >= 2 {
			return cost
		}
	}
	return 0.0
}

func calcLongSFBPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand && curr.associatedFinger == old1.associatedFinger {
		delta := AbsI(curr.row - old1.row)
		if delta >= 2 {
			return cost
		}
	}
	return 0.0
}

func calcLateralStretchPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}

	// Check if both keys were pressed by the same hand and there is a long jump between rows
	if curr.hand == old1.hand {
		// Check if there's a vertical jump between the top and bottom rows
		if (curr.vertDeltaToHome < 0 && old1.vertDeltaToHome > 0) ||
			(curr.vertDeltaToHome > 0 && old1.vertDeltaToHome < 0) {

			// Check for specific finger combinations that are penalized
			if (curr.associatedFinger == Ring && old1.associatedFinger == Pinkie) ||
				(curr.associatedFinger == Pinkie && old1.associatedFinger == Ring) ||
				(curr.associatedFinger == Middle && old1.associatedFinger == Ring) ||
				(curr.associatedFinger == Ring && old1.associatedFinger == Middle) ||
				(curr.associatedFinger == Index &&
					(old1.associatedFinger == Middle || old1.associatedFinger == Ring) &&
					curr.vertDeltaToHome < 0 && old1.vertDeltaToHome > 0) {
				// Apply the penalty
				return cost
			}
		}
	}

	return 0.0

}

func calcPinkyRingStretchPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand {
		if curr.associatedFinger == Pinkie && old1.associatedFinger == Ring {
			if curr.row < old1.row {
				return cost
			}
		} else if curr.associatedFinger == Ring && old1.associatedFinger == Pinkie {
			if curr.row < old1.row {
				return cost
			}
		}
	}
	return 0.0
}

func calcRollReversalPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil || old2 == nil {
		return 0.0
	}
	if curr.hand == old1.hand && old1.hand == old2.hand {
		// Check for a roll reversal where finger sequence reverses
		if (curr.associatedFinger == Middle && old1.associatedFinger == Pinkie && old2.associatedFinger == Ring) ||
			(curr.associatedFinger == Ring && old1.associatedFinger == Pinkie && old2.associatedFinger == Middle) {
			return cost
		}
	}
	return 0.0
}

func calcHandRepetitionPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil || old2 == nil || old3 == nil {
		return 0.0
	}
	// Check if all keys were pressed by the same hand
	if curr.hand == old1.hand && old1.hand == old2.hand && old2.hand == old3.hand {
		return cost
	}
	return 0.0
}

func calcHandAlternationPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil || old2 == nil || old3 == nil {
		return 0.0
	}
	// Check if the hands alternate three times in a row
	if curr.hand != old1.hand && old1.hand != old2.hand && old2.hand != old3.hand {
		return cost
	}
	return 0.0
}

func calcOutwardRollPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand {
		if isRollOut(curr.associatedFinger, old1.associatedFinger) {
			return cost
		}
	}
	return 0.0
}

func calcInwardRollPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand {
		if isRollIn(curr.associatedFinger, old1.associatedFinger) {
			return cost
		}
	}
	return 0.0
}

func calcScissorMotionPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.hand == old1.hand {
		// Penalize scissor-like motion (e.g., ring finger and index finger pressing on opposite rows)
		if (curr.associatedFinger == Ring && old1.associatedFinger == Index && curr.row != old1.row) ||
			(curr.associatedFinger == Index && old1.associatedFinger == Ring && curr.row != old1.row) {
			return cost
		}
	}
	return 0.0
}

func calcRowChangeInRollPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil || old2 == nil {
		return 0.0
	}

	if curr.hand == old1.hand && curr.hand == old2.hand {
		// Check if the row movement spans all three rows (Top -> Home -> Bottom or Bottom -> Home -> Top)
		if (curr.vertDeltaToHome < 0 && old1.vertDeltaToHome == 0 && old2.vertDeltaToHome > 0) ||
			(curr.vertDeltaToHome > 0 && old1.vertDeltaToHome == 0 && old2.vertDeltaToHome < 0) {

			// Check if the movement is a roll out or roll in
			if (isRollOut(curr.associatedFinger, old1.associatedFinger) && isRollOut(old1.associatedFinger, old2.associatedFinger)) ||
				(isRollIn(curr.associatedFinger, old1.associatedFinger) && isRollIn(old1.associatedFinger, old2.associatedFinger)) {

				// Apply the penalty
				return cost
			}
		}
	}
	return 0.0
}

func calcBaseModifierPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if modCurr == nil {
		return 0.0
	}
	return modCurr.cost * cost
}

func calcSameFingerModifierPenalty(curr, old1, old2, old3, modCurr, mod1, mod2, mod3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || modCurr == nil {
		return 0.0
	}
	if curr.hand == modCurr.hand {
		if curr.associatedFinger == modCurr.associatedFinger {
			return cost
		}
	}

	return 0.0
}

// CalculatePenalty calculates the total penalty for a layout and the given quartads.
func CalculatePenalty(quartads QuartadList, layout Layout, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo, penalties []KeyPenalty, detailed bool) (float64, []KeyPenaltyResult) {
	var totalPenalty float64
	results := make([]KeyPenaltyResult, len(penalties))

	for i, penalty := range penalties {
		results[i] = KeyPenaltyResult{
			Name:     penalty.Name,
			Total:    0.0,
			HighKeys: make(map[Quartad]float64),
			Info:     &penalty,
		}
	}

	for quartad, count := range quartads {
		penalty := penalize(quartad, count, layout, runesToKeyPhysicalKeyInfoMap, results, detailed)
		totalPenalty += penalty
	}

	return totalPenalty, results
}

func isRollOut(currFinger, prevFinger Finger) bool {
	switch currFinger {
	case Thumb:
		return false
	case Index:
		return prevFinger == Thumb
	case Middle:
		return prevFinger == Thumb || prevFinger == Index
	case Ring:
		return prevFinger == Middle || prevFinger == Index
	case Pinkie:
		return prevFinger == Ring || prevFinger == Middle
	default:
		return false
	}
}

func isRollIn(currFinger, prevFinger Finger) bool {
	switch currFinger {
	case Thumb:
		return prevFinger != Thumb
	case Index:
		return prevFinger == Thumb || prevFinger == Index
	case Middle:
		return prevFinger == Pinkie || prevFinger == Ring
	case Ring:
		return prevFinger == Pinkie
	case Pinkie:
		return false
	default:
		return false
	}
}

// calculateQuartadPenalty calculates the penalty for a given quartad.
func penalize(quartad Quartad, count int, layout Layout, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo, penalties []KeyPenaltyResult, detailed bool) float64 {
	total := 0.0

	// Get current rune key press information
	curr := getKey(quartad, 0, runesToKeyPhysicalKeyInfoMap)
	old1 := getKey(quartad, 1, runesToKeyPhysicalKeyInfoMap)
	old2 := getKey(quartad, 2, runesToKeyPhysicalKeyInfoMap)
	old3 := getKey(quartad, 3, runesToKeyPhysicalKeyInfoMap)
	modCurr := getModifier(quartad, 0, runesToKeyPhysicalKeyInfoMap)
	mod1 := getModifier(quartad, 0, runesToKeyPhysicalKeyInfoMap)
	mod2 := getModifier(quartad, 0, runesToKeyPhysicalKeyInfoMap)
	mod3 := getModifier(quartad, 0, runesToKeyPhysicalKeyInfoMap)

	for i, penalty := range penalties {
		cost := penalty.Info.Function(curr, old1, old2, old3, modCurr, mod1, mod2, mod3, penalty.Info.Cost) * float64(count)
		total += cost
		if detailed {
			penalties[i].Total += cost
			penalties[i].HighKeys[quartad] += cost
		}
	}

	return total
}

// getKey returns the key press information from the layout.
func getKey(quartad Quartad, reverseIndex int, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo) *KeyPhysicalInfo {
	index := quartad.Len() - (reverseIndex + 1)
	if index <= 0 || reverseIndex < 0 {
		return nil
	}
	r := quartad.GetRune(index)
	return runesToKeyPhysicalKeyInfoMap[r]
}

func getModifier(quartad Quartad, reverseIndex int, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo) *KeyPhysicalInfo {
	index := quartad.Len() - (reverseIndex + 1)
	if index <= 0 || reverseIndex < 0 {
		return nil
	}
	m := rune(quartad.GetModifier(index))
	return runesToKeyPhysicalKeyInfoMap[m]
}
