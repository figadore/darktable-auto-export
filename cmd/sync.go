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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	fmt.Println("sync called")
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
	// Recurse through input directory
	raws := sidecars.FindFilesWithExt(syncOpts.inputPath, syncOpts.extension)
	config := parseConfig()
	for _, raw := range raws {
		fmt.Println(raw)
		// Find adjacent xmp files
		xmps := sidecars.FindXmps(raw)
		basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
		relativeDir := strings.TrimPrefix(filepath.Dir(raw), syncOpts.inputPath)
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
				syncFile(xmp)
			}
		}
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

// syncFile takes the path to a raw file or xmp and exports jpgs
func syncFile(path string) error {
	switch ext := filepath.Ext(syncOpts.inputPath); {
	case ext == ".xmp":
		xmp := path
		fmt.Println("Syncing xmp")
		raw := sidecars.GetRawPathForXmp(xmp, syncOpts.extension)
		basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
		relativeDir := strings.TrimPrefix(filepath.Dir(raw), syncOpts.inputPath)
		outputPath := filepath.Join(syncOpts.outputFolder, relativeDir, fmt.Sprintf("%s.jpg", basename))
		params := darktable.ExportParams{
			Command:    syncOpts.command,
			RawPath:    raw,
			OutputPath: outputPath,
		}
		fmt.Println("  ", xmp)
		// Export the RAW file
		params.XmpPath = xmp
		jpgFilename := sidecars.GetJpgFilename(xmp, syncOpts.extension)
		outputPath, err := filepath.Abs(filepath.Join(syncOpts.outputFolder, relativeDir, jpgFilename))
		if err != nil {
			log.Fatalf("Error getting jpg path: %v", err)
		}
		params.OutputPath = outputPath
		darktable.Export(params)
	case strings.EqualFold(ext, syncOpts.extension):
		fmt.Println("Syncing raw file")
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
