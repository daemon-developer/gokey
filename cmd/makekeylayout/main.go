package main

import (
	"fmt"
)

func main() {
	// Will in the future take the user json and corpus from the command line
	user, err := ReadUser("users/mark.json")
	if err != nil {
		panic(err)
	}
	layout, err := ReadLayout(user)
	if err != nil {
		panic(err)
	}

	quartadInfo, err := GetQuartadList("corpus/aliceinwonderland.txt", layout)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d unique runes\n", len(quartadInfo.RunesToPlace))
	fmt.Printf("%d unused runes\n", len(quartadInfo.RunesNotBeingPlaced))
	fmt.Printf("%d quartads\n", len(quartadInfo.Quartads))

	//// Sort and display top 50 quartads
	//sortedQuartads := sortMapByValueDesc(quartadInfo.Quartads)
	//fmt.Println("Top 50 Quartads:")
	//for i := 0; i < 50 && i < len(sortedQuartads); i++ {
	//	fmt.Printf("%q: %d\n", sortedQuartads[i].Key, sortedQuartads[i].Value)
	//}

	//// Used runes
	//fmt.Println("Used runes:")
	//for _, v := range quartadInfo.RunesToPlace {
	//	fmt.Printf("%q\n", v)
	//}
	//
	//// Unused runes
	//fmt.Println("Unused runes:")
	//for _, v := range quartadInfo.RunesNotBeingPlaced {
	//	fmt.Printf("%q\n", v)
	//}

	fmt.Println(layout.StringWithCosts())

	Optimize(quartadInfo, layout, user, true, 1, 3)
}
