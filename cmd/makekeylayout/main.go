package main

import (
	"fmt"
)

func main() {
	// In the future I will take the user json and corpus from the command line
	user, err := ReadUser("users/mark.json")
	if err != nil {
		panic(err)
	}

	quartadInfo, err := GetQuartadList("corpus/alice-and-prog.txt", user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d runes on keyboard\n", len(quartadInfo.RunesOnKeyboard))
	fmt.Printf("%d quartads\n", len(quartadInfo.Quartads))

	fmt.Println(user.Layout.StringWithCosts())

	Optimize(quartadInfo, user.Layout, user, false, 1, 3)
}
