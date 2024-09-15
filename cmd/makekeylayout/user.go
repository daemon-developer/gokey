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
	Name        string `json:"name"`
	Keyboard    string `json:"keyboard"`
	RawLocale   string `json:"locale"`
	RawRequired string `json:"required"`
	Required    []rune
	Locale      Locale
	Layout      Layout
	Left        Hand `json:"left"`
	Right       Hand `json:"right"`
	Penalties   struct {
		SFB                  float64 `json:"sfb"`
		VerticalFingerTravel float64 `json:"vertical_finger_travel"`
		LongSFB              float64 `json:"long_sfb"`
		LateralStretch       float64 `json:"lateral_stretch"`
		PinkyRingStretch     float64 `json:"pinky_ring_stretch"`
		RollReversal         float64 `json:"roll_reversal"`
		HandRepetition       float64 `json:"hand_repetition"`
		HandAlternation      float64 `json:"hand_alternation"`
		OutwardRoll          float64 `json:"outward_roll"`
		InwardRoll           float64 `json:"inward_roll"`
		ScissorMotion        float64 `json:"scissor_motion"`
		RowChangeInRoll      float64 `json:"row_change_in_roll"`
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

	// Get the required runes
	profile.Required = []rune(profile.RawRequired)

	// Now read their locale
	profile.Locale, err = LoadUserLocale(profile.RawLocale)
	if err != nil {
		return User{}, err
	}

	// Now read their layout
	layout, err := ReadLayout(profile)
	if err != nil {
		return User{}, err
	}
	profile.Layout = layout

	return profile, nil
}
