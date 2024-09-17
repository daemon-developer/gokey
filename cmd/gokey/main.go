package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"time"
)

var (
	r             *rand.Rand
	optIterations int
	optDebug      int
	optSwaps      int
	optLayout     string
	rootCmd       = &cobra.Command{
		Use:   "gokey [username]",
		Short: "Generate a personalized keyboard layout.",
		Long:  `Generate a personalized keyboard layout.`,
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}
)

func init() {
	rootCmd.Flags().IntVarP(&optIterations, "iterations", "i", 10000, "Number of iterations")
	rootCmd.Flags().IntVarP(&optSwaps, "swaps", "s", 3, "Number key swaps per iteration")
	rootCmd.Flags().StringVarP(&optLayout, "layout", "l", "", "Override layout name")
	rootCmd.Flags().IntVarP(&optDebug, "debug", "d", 0, "Debug level (0-2)")
}

func main() {
	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	r = rand.New(source)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	username := args[0]

	// In the future I will take the user json and corpus from the command line
	userConfigFile := "users/" + username + ".json"
	user, err := ReadUser(userConfigFile)
	if err != nil {
		panic(err)
	}

	quartadInfo, err := GetQuartadList(user.Corpus, user)
	if err != nil {
		fmt.Println(err)
		return
	}
	if optDebug > 1 {
		fmt.Printf("%d runes on keyboard\n", len(quartadInfo.RunesOnKeyboard))
		fmt.Printf("%d quartads\n", len(quartadInfo.Quartads))
	}

	fmt.Println(user.Layout.StringWithCosts())

	Optimize(quartadInfo, user.Layout, user, optIterations, optSwaps)
}
