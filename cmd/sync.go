/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
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
	syncCmd.Flags().StringVarP(&syncOpts.inputFolder, "in", "i", "./", "Directory to look for raw image files")
	syncCmd.Flags().StringVarP(&syncOpts.outputFolder, "out", "o", "./", "Directory to export jpgs to")
	syncCmd.Flags().StringVarP(&syncOpts.command, "command", "c", "flatpak run --command=darktable-cli org.darktable.Darktable", "Darktable command or binary")
	syncCmd.Flags().StringVarP(&syncOpts.extension, "extension", "e", ".ARW", "Extension of raw files")
}

type syncOptions struct {
	inputFolder  string
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
	// Recurse through input directory
	raws := sidecars.FindFilesWithExt(syncOpts.inputFolder, syncOpts.extension)
	config := parseConfig()
	for _, raw := range raws {
		fmt.Println(raw)
		// Find adjacent xmp files
		xmps := sidecars.FindXmps(raw)
		basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
		relativeDir := strings.TrimPrefix(filepath.Dir(raw), syncOpts.inputFolder)
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
			}
		}
	}
	// Delete jpgs for missing raws
	if config.DeleteMissing == "true" {
		fmt.Println("Deleting jpgs for missing raws")
		jpgs := sidecars.FindFilesWithExt(syncOpts.outputFolder, ".jpg")
		deleteJpgs(sidecars.FindJpgsWithoutRaw(jpgs, syncOpts.inputFolder, syncOpts.outputFolder, syncOpts.extension))
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

func deleteJpgs(jpgs []string) {
	for _, jpgPath := range jpgs {
		err := os.Remove(jpgPath)
		if err != nil {
			log.Fatalf("Error deleting jpg: %v", err)
			return
		}
	}
}
