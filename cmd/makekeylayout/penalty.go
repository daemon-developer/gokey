package main

// KeyPenalty defines a penalty rule.
type KeyPenalty struct {
	Name     string
	Function PenaltyFunc
	Cost     float64
}

type PenaltyFunc func(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64

type KeyPenaltyResult struct {
	Name     string
	Total    float64
	HighKeys map[string]float64
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
	}
}

func calcBasePenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil {
		return 0.0
	}
	return curr.cost * cost
}

func calcSFBPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil || old1 == nil {
		return 0.0
	}
	if curr.associatedFinger == old1.associatedFinger {
		if curr.key != old1.key {
			return cost
		}
	}
	return 0.0
}

func calcVerticalFingerTravelPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcLongSFBPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcLateralStretchPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcPinkyRingStretchPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcRollReversalPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcHandRepetitionPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcHandAlternationPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcOutwardRollPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcInwardRollPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcScissorMotionPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func calcRowChangeInRollPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
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
			HighKeys: make(map[string]float64),
			Info:     &penalty,
		}
	}

	for quartad, count := range quartads {
		penalty := penalize(quartad, count, layout, runesToKeyPhysicalKeyInfoMap, results, detailed)
		totalPenalty += penalty
	}

	return totalPenalty, results
}

// calculateQuartadPenalty calculates the penalty for a given quartad.
func penalize(quartad string, count int, layout Layout, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo, penalties []KeyPenaltyResult, detailed bool) float64 {
	total := 0.0

	// Get current rune key press information
	curr := getKey(quartad, 0, runesToKeyPhysicalKeyInfoMap)
	old1 := getKey(quartad, 1, runesToKeyPhysicalKeyInfoMap)
	old2 := getKey(quartad, 2, runesToKeyPhysicalKeyInfoMap)
	old3 := getKey(quartad, 3, runesToKeyPhysicalKeyInfoMap)

	for i, penalty := range penalties {
		cost := penalty.Info.Function(curr, old1, old2, old3, penalty.Info.Cost) * float64(count)
		total += cost
		if detailed {
			penalties[i].Total += cost
			penalties[i].HighKeys[quartad] += cost
		}
	}

	return total
}

// getKey returns the key press information from the layout.
func getKey(quartad string, reverseIndex int, runesToKeyPhysicalKeyInfoMap map[rune]*KeyPhysicalInfo) *KeyPhysicalInfo {
	index := len(quartad) - (reverseIndex + 1)
	if index <= 0 || reverseIndex < 0 {
		return nil
	}
	r := rune(quartad[index])
	return runesToKeyPhysicalKeyInfoMap[r]
}
