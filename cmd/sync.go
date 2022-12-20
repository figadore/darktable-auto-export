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
	"github.com/spf13/viper"
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

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Local flags which will only run when this command is called directly
	syncCmd.Flags().StringP("in", "i", "./", "Directory or file of raw image(s)")
	syncCmd.Flags().StringP("out", "o", "./", "Directory to export jpgs to")
	syncCmd.Flags().StringP("command", "c", "flatpak run --command=darktable-cli org.darktable.Darktable", "Darktable command or binary")
	syncCmd.Flags().StringSliceP("extension", "e", []string{".ARW"}, "Extension of raw files")
	syncCmd.Flags().BoolP("new", "n", false, "Only export when target jpg does not exist")
	syncCmd.Flags().BoolP("delete-missing", "d", false, `Delete jpgs where corresponding raw files are missing. This is useful for darktable workflows where editing and culling can be done at any time, not just up front. *warning* This will delete all jpgs in the output directory where a corresponding raw file with the specified extension cannot be found! Only use this for directories that are exclusively for this workflow, and where the source files stay where they are/were.
`)

	viper.SetConfigName("config")
	// Is viper.SetConfigType() needed here?
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.darktable-auto-export")
	err := viper.ReadInConfig()
	if err != nil {
		// Only allow config file not found error
		// Not sure why viper.ConfigFileNotFoundError doesn't work with errors.Is() or errors.As()
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("Fatal error reading config file: %w", err))
		}
	}
	viper.BindPFlags(syncCmd.Flags())
}

func sync(cmd *cobra.Command, args []string) error {
	// Check whether input arg is a directory or a xmp file
	isDir, err := sidecars.IsDir(viper.GetString("in"))
	if err != nil {
		return err
	}

	if isDir {
		return syncDir()
	} else {
		return syncFile(viper.GetString("in"))
	}

}

func syncDir() error {
	// Recurse through input directory looking for finding raw files
	var raws []string
	// For each defined extension, add to the list of raws
	for _, ext := range viper.GetStringSlice("extension") {
		raws = append(raws, sidecars.FindFilesWithExt(viper.GetString("in"), ext)...)
	}
	for _, raw := range raws {
		err := syncRaw(raw)
		if err != nil {
			return err
		}
	}
	// Delete jpgs with missing raws and xmps
	if viper.GetBool("delete-missing") {
		fmt.Println("Deleting jpgs for missing raws")
		jpgs := sidecars.FindFilesWithExt(viper.GetString("out"), ".jpg")
		jpgsToDelete := sidecars.FindJpgsWithoutRaw(jpgs, raws, viper.GetString("in"), viper.GetString("out"), viper.GetStringSlice("extension"))
		jpgsToDelete = append(jpgsToDelete, sidecars.FindJpgsWithoutXmp(jpgs, viper.GetString("in"), viper.GetString("out"), viper.GetStringSlice("extension"))...)
		deleteJpgs(jpgsToDelete)
		fmt.Printf("Deleting %v of %v jpgs", len(jpgsToDelete), len(jpgs))
	} else {
		fmt.Printf("Not deleting jpgs for missing raws")
	}
	// Look for xmp file(s) for the raw file
	// If no xmp file exists for a RAW...
	// Run darktable cli, setting export path to match structure of input dir
	//  darktable-cli [<input file or dir>] [<xmp file>] <output destination> [options] [--core <darktable options>]
	fmt.Println("\nComplete")
	return nil
}

func syncRaw(raw string) error {
	fmt.Println("Syncing raw", raw)
	// Find adjacent xmp files
	xmps := sidecars.FindXmps(raw)
	basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
	relativeDir := sidecars.GetRelativeDir(raw, viper.GetString("in"))
	outputPath := filepath.Join(viper.GetString("out"), relativeDir, fmt.Sprintf("%s.jpg", basename))
	params := darktable.ExportParams{
		Command:    viper.GetString("command"),
		RawPath:    raw,
		OutputPath: outputPath,
		OnlyNew:    viper.GetBool("new"),
	}
	if len(xmps) == 0 {
		fmt.Println("No xmp files found, applying default settings")
		err := darktable.Export(params)
		if err != nil {
			return err
		}
	} else {
		for _, xmp := range xmps {
			fmt.Println("Sync xmp file", xmp)
			err := syncFile(xmp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// syncFile takes the path to a raw file or xmp and exports jpgs
func syncFile(path string) error {
	//switch ext := filepath.Ext(viper.GetString("in")); {
	switch ext := filepath.Ext(path); {
	case ext == ".xmp":
		xmp := path
		fmt.Println("Syncing xmp")
		foundRaw, err := sidecars.FindRawPathForXmp(xmp, viper.GetStringSlice("extension"))
		if err != nil {
			return err
		}
		relativeDir := sidecars.GetRelativeDir(xmp, viper.GetString("in"))
		// Export the RAW file
		jpgFilename := sidecars.GetJpgFilename(xmp, viper.GetStringSlice("extension"))
		outputPath, err := filepath.Abs(filepath.Join(viper.GetString("out"), relativeDir, jpgFilename))
		if err != nil {
			log.Fatalf("Error getting jpg path: %v", err)
		}
		params := darktable.ExportParams{
			Command:    viper.GetString("command"),
			RawPath:    foundRaw,
			OnlyNew:    viper.GetBool("new"),
			OutputPath: outputPath,
			XmpPath:    xmp,
		}
		err = darktable.Export(params)
		if err != nil {
			return err
		}
	// raw
	case caseInsensitiveContains(viper.GetStringSlice("extension"), ext):
		fmt.Println("Syncing raw file with extension", ext, ":", path)
		err := syncRaw(path)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Extension of file to be synced ('%s') does not match the extension specified for processing ('%s')", ext, viper.GetStringSlice("extension")))
	}
	return nil
}

func caseInsensitiveContains(haystack []string, needle string) bool {
	fmt.Println("Checking", haystack, "for", needle)
	for _, v := range haystack {
		if strings.EqualFold(needle, v) {
			fmt.Println("found")
			return true
		}
	}
	fmt.Println("not found")
	return false
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
