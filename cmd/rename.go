package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/merensoft/merenpic/models"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Run:   runRenameCommand,
	Use:   "rename <folder-path>",
	Short: "rename allow you to unify your media files naming",
	Long: "Automatically rename your media files to a unified format, " +
		"only run this after photos are grouped otherwise this " +
		"would have a hard time finding the photo files.",
}

var renameFlags struct {
	folder string
}

func init() {
	rootCmd.AddCommand(renameCmd)

	renameCmd.Flags().StringVar(&renameFlags.folder, "folder", "", "folder path to be process")
}

func runRenameCommand(cmd *cobra.Command, _ []string) {
	if renameFlags.folder == "" {
		_ = cmd.Help()
		return
	}

	fmt.Printf("Running Rename Command for folder: %s\n", renameFlags.folder)

	// check the folder path exists
	if _, err := os.Stat(renameFlags.folder); os.IsNotExist(err) {
		PrintErrorAndExit(fmt.Errorf("folder path does not exist: %s", renameFlags.folder))
	}

	// find all the json files including the subdirectories
	nFiles := countFiles(renameFlags.folder)

	bar := progressbar.NewOptions(nFiles,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(100),
		progressbar.OptionSetSpinnerChangeInterval(10*time.Millisecond),
		progressbar.OptionSetDescription("Renaming files..."),
	)

	renameFilesRecursively(renameFlags.folder, bar)

	ExitIfError(bar.Finish())

	fmt.Println("\nAll files renamed successfully!")
}

func countFiles(folder string) int {
	count := 0

	_ = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// Skip files with ._ prefix
		if strings.HasPrefix(path, "._") {
			return nil
		}

		if strings.HasSuffix(path, ".json") {
			count++
		}

		return nil
	})

	return count
}

func renameFilesRecursively(folder string, bar *progressbar.ProgressBar) {
	allFiles, err := os.ReadDir(folder)
	ExitIfError(err)

	for _, dirEntry := range allFiles {
		if dirEntry.IsDir() {
			renameFilesRecursively(fmt.Sprintf("%s/%s", folder, dirEntry.Name()), bar)
		} else {
			if strings.HasSuffix(dirEntry.Name(), ".json") {
				readJsonAndRenameFiles(folder, dirEntry.Name(), bar)
			}
		}
	}
}

func readJsonAndRenameFiles(folderPath string, fileName string, bar *progressbar.ProgressBar) {
	// read the json file
	filePath := fmt.Sprintf("%s/%s", folderPath, fileName)

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
			return
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

	// format the new file name, example: IMG_2021_01_01_123055_001.jpg
	subfix := 0
	newPhotoName := ""
	newPhotoPath := ""
	prefix, fileType := getFileType(metadata.Title)
	day := photoTime.Day()
	year := photoTime.Year()
	month := int(photoTime.Month())
	timeValue := fmt.Sprintf("%02d%02d%02d", photoTime.Hour(), photoTime.Minute(), photoTime.Second())

	for subfix < 1000 {
		subfix++
		newPhotoName = fmt.Sprintf("%s_%d_%02d_%02d_%s_%03d.%s", prefix, year, month, day, timeValue, subfix, fileType)
		newPhotoPath = fmt.Sprintf("%s/%s", folderPath, newPhotoName)
		if FileExists(newPhotoPath) {
			// file exists, try the next subfix
			continue
		}

		break
	}

	// move the media file
	oldPhotoPath := fmt.Sprintf("%s/%s", folderPath, strings.TrimSuffix(fileName, ".json"))
	ExitIfError(os.Rename(oldPhotoPath, newPhotoPath))

	// rename the json file
	newJsonName := fmt.Sprintf("%s.json", newPhotoPath)
	ExitIfError(os.Rename(filePath, newJsonName))

	ExitIfError(bar.Add(1))
}

func getFileType(fileName string) (string, string) {
	fileType := strings.ToLower(filepath.Ext(fileName))

	switch fileType {
	case ".jpg", ".jpeg":
		return "IMG", "jpg"
	case ".png":
		return "IMG", "png"
	case ".mp4":
		return "VID", "mp4"
	case ".mov":
		return "VID", "mov"
	case ".gif":
		return "GIF", "gif"
	default:
		// unknown file type return filetype without . prefix
		return "UNK", strings.TrimSuffix(fileType, ".")
	}
}
