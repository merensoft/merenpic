package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "merenpic",
	Version: "0.0.1",
	Short:   "Merenpic CLI tool",
	Long:    `Merenpic is a CLI tool that help you organize your photos and videos from a Google Photos takeout.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// shared functions

func ExitIfError(err error) {
	if err != nil {
		PrintErrorAndExit(err)
	}
}

func PrintErrorAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
