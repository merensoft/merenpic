package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	jsonFiles, err := walkDir(groupFlags.folder, "json")
	ExitIfError(err)

	fmt.Printf("Found %d json files in the folder\n", len(jsonFiles))

	bar := progressbar.NewOptions(len(jsonFiles),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(100),
		progressbar.OptionSetSpinnerChangeInterval(10*time.Millisecond),
		progressbar.OptionSetDescription("Moving files..."),
	)

	filesWithError := make([]string, 0)

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

		photoTime := time.Unix(unixTime, 0).UTC()

		// create the subdirectory if it does not exist
		subDirName := fmt.Sprintf("%s/photos_%d_%02d", groupFlags.folder, photoTime.Year(), int(photoTime.Month()))

		if _, err := os.Stat(subDirName); os.IsNotExist(err) {
			err = os.Mkdir(subDirName, 0755)
			ExitIfError(err)
		}

		// check if the photo file exists
		photoName := strings.TrimSuffix(file, ".json")
		photoFilePath := fmt.Sprintf("%s/%s", groupFlags.folder, photoName)
		if fileExists(photoFilePath) {
			moveFiles(subDirName, filePath, photoFilePath, photoName)
			ExitIfError(bar.Add(1))
			continue
		}

		// edge case where the photo name is longer that the metadata file name
		// example PXL_123.LONG_EXPOSURE-02.ORIGINA.jpg and PXL_123.LONG_EXPOSURE-02.ORIGIN.json
		// in this case we should search for a photo fie that start with the metadata file name
		fileFinds, err := filepath.Glob(fmt.Sprintf("%s/%s*", groupFlags.folder, photoName))
		if err == nil && len(fileFinds) == 2 {
			// we found the photo file
			photoName = metadata.Title

			if strings.Contains(fileFinds[0], ".json") {
				photoFilePath = fileFinds[1]
			} else {
				photoFilePath = fileFinds[0]
			}

			moveFiles(subDirName, filePath, photoFilePath, photoName)
			ExitIfError(bar.Add(1))
			continue
		}

		// edge case where the photo name has (n) suffix
		// example photo(1).jpg and photo.jpg(n).json
		// we should be able to handle this case, by moving the photo with (n) suffix
		// and moving json with better suffix: photo(n).jpg.json
		re := regexp.MustCompile(`\((\d+)\)\.json$`)
		matches := re.FindStringSubmatch(file)
		if len(matches) > 0 {
			suffix := matches[1]
			photoName = strings.TrimSuffix(file, fmt.Sprintf("(%s).json", suffix))
			splitName := strings.Split(photoName, ".")
			if len(splitName) != 2 {
				filesWithError = append(filesWithError, file)
				continue
			}

			photoName = fmt.Sprintf("%s(%s).%s", splitName[0], suffix, splitName[1])
			photoFilePath = fmt.Sprintf("%s/%s", groupFlags.folder, photoName)

			if _, err := os.Stat(photoFilePath); os.IsNotExist(err) {
				filesWithError = append(filesWithError, file)
				continue
			}

			moveFiles(subDirName, filePath, photoFilePath, photoName)
			ExitIfError(bar.Add(1))
			continue
		}

		// another edge case, where the photo file exists with the title
		photoName = metadata.Title
		photoFilePath = fmt.Sprintf("%s/%s", groupFlags.folder, photoName)
		if fileExists(photoFilePath) {
			moveFiles(subDirName, filePath, photoFilePath, photoName)
			ExitIfError(bar.Add(1))
			continue
		}

		// if we reach this point, we could not find the photo file
		filesWithError = append(filesWithError, file)

		ExitIfError(bar.Add(1))
	}

	ExitIfError(bar.Finish())

	fmt.Println("\nGroup Command Completed")

	if len(filesWithError) > 0 {
		fmt.Println("Files with error:")
		for _, file := range filesWithError {
			fmt.Println(file)
		}
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func moveFiles(subDirName string, filePath string, photoFilePath string, photoName string) {
	// move the photo file to the subdirectory
	err := os.Rename(photoFilePath, fmt.Sprintf("%s/%s", subDirName, photoName))
	ExitIfError(err)

	// move the metadata file to the subdirectory if photo file was moved
	err = os.Rename(filePath, fmt.Sprintf("%s/%s.json", subDirName, photoName))
	ExitIfError(err)
}

func walkDir(root string, extension string) ([]string, error) {
	files := make([]string, 0)

	allFiles, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range allFiles {
		if file.IsDir() {
			continue
		}

		// Skip files with ._ prefix
		if strings.HasPrefix(file.Name(), "._") {
			continue
		}

		if strings.HasSuffix(file.Name(), "."+extension) {
			files = append(files, file.Name())
		}
	}

	return files, err
}
