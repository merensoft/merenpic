package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/merensoft/merenpic/models"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Run:   runGroupCommand,
	Use:   "group <folder-path>",
	Short: "group move your photos and videos to subdirectories",
	Long:  `Automatically move your media to subdirectories based on the google metadata files.`,
}

var groupFlags struct {
	folder string
}

func init() {
	rootCmd.AddCommand(groupCmd)

	groupCmd.Flags().StringVar(&groupFlags.folder, "folder", "", "folder path to be process")
}

func runGroupCommand(cmd *cobra.Command, _ []string) {
	if groupFlags.folder == "" {
		_ = cmd.Help()
		return
	}

	fmt.Printf("Running Group Command for folder: %s\n", groupFlags.folder)

	// check the folder path exists
	if _, err := os.Stat(groupFlags.folder); os.IsNotExist(err) {
		PrintErrorAndExit(fmt.Errorf("folder path does not exist: %s", groupFlags.folder))
	}

	// find all the files in the folder
	files, err := os.ReadDir(groupFlags.folder)
	ExitIfError(err)

	// filter out only the json files
	var jsonFiles []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		jsonFiles = append(jsonFiles, file.Name())
	}

	fmt.Printf("Found %d json files in the folder\n", len(jsonFiles))

	bar := progressbar.NewOptions(len(jsonFiles),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(100),
		progressbar.OptionSetSpinnerChangeInterval(10*time.Millisecond),
		progressbar.OptionSetDescription("Moving files..."),
	)

	// open each json file and read the metadata, the photo taken photoTakenTime and move the file to the subdirectory
	// with the following format: photos-<year>-<month> e.g. photos-2021-01
	for _, file := range jsonFiles {
		filePath := fmt.Sprintf("%s/%s", groupFlags.folder, file)

		// open file
		fileText, err := os.ReadFile(filePath)
		ExitIfError(err)

		// parse the json file
		var metadata models.PhotoMetadata

		err = json.Unmarshal(fileText, &metadata)
		ExitIfError(err)

		var photoTakenTime string

		if metadata.PhotoTakenTime == nil {
			if metadata.CreationTime.Timestamp == "" {
				fmt.Println("\nNo photo taken time found, skipping...")
				continue
			} else {
				photoTakenTime = metadata.CreationTime.Timestamp
			}
		} else {
			photoTakenTime = metadata.PhotoTakenTime.Timestamp
		}

		// get the photo taken time
		unixTime, err := strconv.ParseInt(photoTakenTime, 10, 64)
		ExitIfError(err)

		photoTime := time.Unix(unixTime, 0)

		// create the subdirectory if it does not exist
		subDirName := fmt.Sprintf("%s/photos_%d_%02d", groupFlags.folder, photoTime.Year(), int(photoTime.Month()))

		if _, err := os.Stat(subDirName); os.IsNotExist(err) {
			err = os.Mkdir(subDirName, 0755)
			ExitIfError(err)
		}

		// check if the photo file exists
		photoName := strings.TrimSuffix(file, ".json")
		photoFilePath := fmt.Sprintf("%s/%s", groupFlags.folder, photoName)

		if _, err := os.Stat(photoFilePath); os.IsNotExist(err) {
			fmt.Printf("\nPhoto file: %s does not exist, skipping...\n", photoFilePath)
			continue
		}

		// move the photo file to the subdirectory
		err = os.Rename(photoFilePath, fmt.Sprintf("%s/%s", subDirName, photoName))
		ExitIfError(err)

		// move the metadata file to the subdirectory if photo file was moved
		err = os.Rename(filePath, fmt.Sprintf("%s/%s", subDirName, file))
		ExitIfError(err)

		err = bar.Add(1)
		ExitIfError(err)
	}

	fmt.Println("\nGroup Command Completed")
}
