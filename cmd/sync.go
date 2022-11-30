package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/figadore/darktable-auto-export/internal/darktable"
	"github.com/figadore/darktable-auto-export/internal/sidecars"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Export jpgs for raw files",
	Long: `Export jpgs for raw files
Given a directory with just raw files, a jpg will be produced for each raw
Given a directory with raw and xmp files, a jpg will be produced for each raw/xmp combination
Given a raw file, a jpg will be produced for each raw/xmp combination
Given a xmp file, a single jpg will be produced for the matching raw`,
	RunE: sync,
}

var syncOpts syncOptions

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags which will only run when this command is called directly
	syncCmd.Flags().StringVarP(&syncOpts.inputPath, "in", "i", "./", "Directory or file of raw image(s)")
	syncCmd.Flags().StringVarP(&syncOpts.outputFolder, "out", "o", "./", "Directory to export jpgs to")
	syncCmd.Flags().StringVarP(&syncOpts.command, "command", "c", "flatpak run --command=darktable-cli org.darktable.Darktable", "Darktable command or binary")
	syncCmd.Flags().StringVarP(&syncOpts.extension, "extension", "e", ".ARW", "Extension of raw files")
}

type syncOptions struct {
	inputPath    string
	outputFolder string
	extension    string
	command      string
}

type Config struct {
	DeleteMissing string `yaml:"delete-missing"`
}

func parseConfig() Config {
	config := Config{}
	ymlContents, err := os.ReadFile("config.yml")
	fmt.Println("yaml contents: ", string(ymlContents))
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	err = yaml.Unmarshal(ymlContents, &config)
	fmt.Println("config: ", config)
	if err != nil {
		log.Fatalf("yaml read error: %v", err)
	}
	return config
}

func sync(cmd *cobra.Command, args []string) error {
	// Check whether input arg is a directory or an xmp file
	file, err := os.Open(syncOpts.inputPath)
	if err != nil {
		return err
	}

	defer file.Close()
	// This returns an *os.FileInfo type
	fileInfo, err := file.Stat()
	if err != nil {
		// error handling
	}

	// IsDir is short for fileInfo.Mode().IsDir()
	if fileInfo.IsDir() {
		return syncDir()
	} else {
		// not a directory
		return syncFile(syncOpts.inputPath)
	}

}

func syncDir() error {
	// Recurse through input directory looking for finding raw files
	raws := sidecars.FindFilesWithExt(syncOpts.inputPath, syncOpts.extension)
	config := parseConfig()
	for _, raw := range raws {
		syncRaw(raw)
	}
	// Delete jpgs for missing raws
	if config.DeleteMissing == "true" {
		fmt.Println("Deleting jpgs for missing raws")
		jpgs := sidecars.FindFilesWithExt(syncOpts.outputFolder, ".jpg")
		deleteJpgs(sidecars.FindJpgsWithoutRaw(jpgs, syncOpts.inputPath, syncOpts.outputFolder, syncOpts.extension))
		fmt.Printf("Found %v jpgs", len(jpgs))
	} else {
		fmt.Printf("Not deleting jpgs for missing raws: %s", config.DeleteMissing)
	}
	// Look for xmp file(s) for the raw file
	// If no xmp file exists for a RAW...
	// Run darktable cli, setting export path to match structure of input dir
	//  darktable-cli [<input file or dir>] [<xmp file>] <output destination> [options] [--core <darktable options>]
	fmt.Println("\nComplete")
	return nil
}

func syncRaw(raw string) {
	fmt.Println("Syncing raw", raw)
	// Find adjacent xmp files
	xmps := sidecars.FindXmps(raw)
	basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
	relativeDir := sidecars.GetRelativeDir(raw, syncOpts.inputPath)
	outputPath := filepath.Join(syncOpts.outputFolder, relativeDir, fmt.Sprintf("%s.jpg", basename))
	params := darktable.ExportParams{
		Command:    syncOpts.command,
		RawPath:    raw,
		OutputPath: outputPath,
	}
	if len(xmps) == 0 {
		fmt.Println("No xmp files found, applying default settings")
		darktable.Export(params)
	} else {
		for _, xmp := range xmps {
			fmt.Println("Sync xmp file", xmp)
			syncFile(xmp)
		}
	}
}

// syncFile takes the path to a raw file or xmp and exports jpgs
func syncFile(path string) error {
	//switch ext := filepath.Ext(syncOpts.inputPath); {
	switch ext := filepath.Ext(path); {
	case ext == ".xmp":
		xmp := path
		fmt.Println("Syncing xmp")
		raw := sidecars.GetRawPathForXmp(xmp, syncOpts.extension)
		basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
		relativeDir := sidecars.GetRelativeDir(raw, syncOpts.inputPath)
		outputPath := filepath.Join(syncOpts.outputFolder, relativeDir, fmt.Sprintf("%s.jpg", basename))
		params := darktable.ExportParams{
			Command:    syncOpts.command,
			RawPath:    raw,
			OutputPath: outputPath,
		}
		// Export the RAW file
		params.XmpPath = xmp
		jpgFilename := sidecars.GetJpgFilename(xmp, syncOpts.extension)
		outputPath, err := filepath.Abs(filepath.Join(syncOpts.outputFolder, relativeDir, jpgFilename))
		if err != nil {
			log.Fatalf("Error getting jpg path: %v", err)
		}
		params.OutputPath = outputPath
		darktable.Export(params)
	// raw
	case strings.EqualFold(ext, syncOpts.extension):
		fmt.Println("Syncing raw file with extension", ext, ":", path)
		syncRaw(path)
	default:
		return errors.New(fmt.Sprintf("Extension of file to be synced ('%s') does not match the extension specified for processing ('%s')", ext, syncOpts.extension))
	}
	return nil
}

func deleteJpgs(jpgs []string) {
	for _, jpgPath := range jpgs {
		err := os.Remove(jpgPath)
		if err != nil {
			log.Fatalf("Error deleting jpg: %v", err)
			return
		}
	}
}
