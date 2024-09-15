package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type FingerCost struct {
	Cost     float64 `json:"cost"`
	UpCost   float64 `json:"up_cost"`
	DownCost float64 `json:"down_cost"`
	HCost    float64 `json:"h_cost"`
}

type Hand struct {
	Thumb  FingerCost `json:"thumb"`
	Index  FingerCost `json:"index"`
	Middle FingerCost `json:"middle"`
	Ring   FingerCost `json:"ring"`
	Pinkie FingerCost `json:"pinkie"`
}

type User struct {
	Name      string `json:"name"`
	Keyboard  string `json:"keyboard"`
	RawLocale string `json:"locale"`
	Locale    Locale
	Left      Hand `json:"left"`
	Right     Hand `json:"right"`
	Penalties struct {
		SameFinger          float64 `json:"same_finger"`
		LongJumpHand        float64 `json:"long_jump_hand"`
		LongJump            float64 `json:"long_jump"`
		LongJumpConsecutive float64 `json:"long_jump_consecutive"`
		PinkyRingTwist      float64 `json:"pinky_ring_twist"`
		RollReversal        float64 `json:"roll_reversal"`
		SameHand            float64 `json:"same_hand"`
		AlternatingHand     float64 `json:"alternating_hand"`
		RollOut             float64 `json:"roll_out"`
		RollIn              float64 `json:"roll_in"`
		LongJumpSandwich    float64 `json:"long_jump_sandwich"`
		Twist               float64 `json:"twist"`
	} `json:"penalties"`
}

func ReadUser(filename string) (User, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return User{}, fmt.Errorf("error reading file: %w", err)
	}

	// Parse JSON
	var profile User
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return User{}, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Now read their locale
	profile.Locale, err = LoadUserLocale(profile.RawLocale)
	if err != nil {
		return User{}, err
	}

	return profile, nil
}
