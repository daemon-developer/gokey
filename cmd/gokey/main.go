package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	optIterations int
	optDebug      int
	optSwaps      int
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
	rootCmd.Flags().IntVarP(&optDebug, "debug", "d", 0, "Debug level (0-2)")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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

	x := len(user.Layout.GetSwappableKeys())
	iterations := float64(x*x) * math.Log(float64(x))
	fmt.Printf("Number of swappable keys: %d\n", x)
	fmt.Printf("Number of recommended iterations: %d\n", int(iterations))
	fmt.Printf("Number of recommended swaps: %d\n", x/10)

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
