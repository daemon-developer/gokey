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
		{Name: "base", Function: BasePenalty, Cost: 0.0},
		{Name: "same finger", Function: SameFingerPenalty, Cost: user.Penalties.SameFinger},
		{Name: "long jump hand", Function: LongJumpHandPenalty, Cost: user.Penalties.LongJumpHand},
		{Name: "long jump", Function: LongJumpPenalty, Cost: user.Penalties.LongJump},
		{Name: "long jump consecutive", Function: LongJumpConsecutivePenalty, Cost: user.Penalties.LongJumpConsecutive},
		{Name: "pinky/ring twist", Function: PinkyRingTwistPenalty, Cost: user.Penalties.PinkyRingTwist},
		{Name: "roll reversal", Function: RollReversalPenalty, Cost: user.Penalties.RollReversal},
		{Name: "same hand", Function: SameHandPenalty, Cost: user.Penalties.SameHand},
		{Name: "alternating hand", Function: AlternatingHandPenalty, Cost: user.Penalties.AlternatingHand},
		{Name: "roll out", Function: RollOutPenalty, Cost: user.Penalties.RollOut},
		{Name: "roll in", Function: RollInPenalty, Cost: user.Penalties.RollIn},
		{Name: "long jump sandwich", Function: LongJumpSandwichPenalty, Cost: user.Penalties.LongJumpSandwich},
		{Name: "twist", Function: TwistPenalty, Cost: user.Penalties.Twist},
	}
}

func BasePenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	if curr == nil {
		return 0.0
	}
	return curr.key.Cost
}

func SameFingerPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func LongJumpHandPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func LongJumpPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func LongJumpConsecutivePenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func PinkyRingTwistPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func RollReversalPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func SameHandPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func AlternatingHandPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func RollOutPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func RollInPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func LongJumpSandwichPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
	// TODO: Implement
	return 0.0
}

func TwistPenalty(curr, old1, old2, old3 *KeyPhysicalInfo, cost float64) float64 {
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
