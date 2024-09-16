package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var (
	optIterations  int
	optSwaps       int
	optDebug       bool
	optParallelism int
	rootCmd        = &cobra.Command{
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
	rootCmd.Flags().IntVarP(&optParallelism, "parallelism", "p", 1, "Number of parallel processes")
	rootCmd.Flags().BoolVarP(&optDebug, "debug", "d", false, "Enable debug mode")
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

	quartadInfo, err := GetQuartadList(user.Corpus, user)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%d runes on keyboard\n", len(quartadInfo.RunesOnKeyboard))
	fmt.Printf("%d quartads\n", len(quartadInfo.Quartads))

	fmt.Println(user.Layout.StringWithCosts())

	Optimize(quartadInfo, user.Layout, user, optDebug, optIterations, 1, optSwaps, optParallelism)
}
