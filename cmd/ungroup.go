package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var ungroupCmd = &cobra.Command{
	Run:   runUngroupCommand,
	Use:   "ungroup <folder-path>",
	Short: "ungroup move your photos and videos from subdirectories",
	Long:  `Automatically move your media from subdirectories to the root folder.`,
}

var ungroupFlags struct {
	folder string
}

func init() {
	rootCmd.AddCommand(ungroupCmd)

	ungroupCmd.Flags().StringVar(&ungroupFlags.folder, "folder", "", "folder path to be process")
}

func runUngroupCommand(cmd *cobra.Command, _ []string) {
	if ungroupFlags.folder == "" {
		_ = cmd.Help()
		return
	}

	fmt.Printf("Running Ungroup Command for folder: %s\n", ungroupFlags.folder)

	// check the folder path exists
	if _, err := os.Stat(ungroupFlags.folder); os.IsNotExist(err) {
		PrintErrorAndExit(fmt.Errorf("folder path does not exist: %s", ungroupFlags.folder))
	}

	allFiles, err := os.ReadDir(ungroupFlags.folder)
	ExitIfError(err)

	bar := progressbar.NewOptions(len(allFiles),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(100),
		progressbar.OptionSetSpinnerChangeInterval(10*time.Millisecond),
		progressbar.OptionSetDescription("Moving files..."),
	)

	for _, dirEntry := range allFiles {
		if dirEntry.IsDir() {
			dirFiles, err := os.ReadDir(fmt.Sprintf("%s/%s", ungroupFlags.folder, dirEntry.Name()))
			ExitIfError(err)

			for _, dirFile := range dirFiles {
				toPath := fmt.Sprintf("%s/%s", ungroupFlags.folder, dirFile.Name())
				fromPath := fmt.Sprintf("%s/%s/%s", ungroupFlags.folder, dirEntry.Name(), dirFile.Name())

				err := os.Rename(fromPath, toPath)
				ExitIfError(err)
			}

			// remove the empty directory
			err = os.Remove(fmt.Sprintf("%s/%s", ungroupFlags.folder, dirEntry.Name()))
			ExitIfError(err)

			err = bar.Add(1)
			ExitIfError(err)
		}
	}

	ExitIfError(bar.Finish())

	fmt.Println("Ungroup command completed")
}
