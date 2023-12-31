package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/figadore/darktable-auto-export/internal/darktable"
	"github.com/figadore/darktable-auto-export/internal/linkedimage"

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
	syncCmd.Flags().Bool("dry-run", false, "Show actions that would be performed, but don't do them")
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
	isDir, err := linkedimage.IsDir(viper.GetString("in"))
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
	inDir := viper.GetString("in")
	outDir := viper.GetString("out")
	extensions := viper.GetStringSlice("extension")
	raws, _, jpgs := linkedimage.FindImages(inDir, outDir, extensions)
	f, err := os.Create("runtime.prof")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	runtime.SetCPUProfileRate(1000)
	pprof.StartCPUProfile(f)
	fmt.Println("starting cpu profile")
	defer pprof.StopCPUProfile()
	for _, raw := range raws {
		params := darktable.ExportParams{
			Command: viper.GetString("command"),
			RawPath: raw.GetPath(),
			OnlyNew: viper.GetBool("new"),
			DryRun:  viper.GetBool("dry-run"),
		}
		err := raw.Sync(params, viper.GetString("out"))
		if err != nil {
			return err
		}
	}
	// Delete jpgs with missing raws and xmps
	if viper.GetBool("delete-missing") {
		fmt.Println("Deleting jpgs for missing raws")
		// Use a map to avoid duplicates
		jpgsToDelete := make(map[string]*linkedimage.Jpg)
		for _, jpg := range jpgs {
			if jpg.Raw == nil {
				jpgsToDelete[jpg.GetPath()] = jpg
			}
			if jpg.Xmp == nil && jpg.IsVirtualCopy() {
				jpgsToDelete[jpg.GetPath()] = jpg
			}
		}
		fmt.Printf("Deleting %v of %v jpgs\n", len(jpgsToDelete), len(jpgs))
		for _, v := range jpgsToDelete {
			v.Delete(viper.GetBool("dry-run"))
		}
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

// syncFile takes the path to a raw file or xmp and exports jpgs
func syncFile(path string) error {
	inDir := viper.GetString("in")
	outDir := viper.GetString("out")
	extensions := viper.GetStringSlice("extension")
	//switch ext := filepath.Ext(viper.GetString("in")); {
	switch ext := filepath.Ext(path); {
	case ext == ".xmp":
		xmp, err := linkedimage.FindXmp(path, inDir, outDir, extensions)
		if err != nil {
			return err
		}
		params := darktable.ExportParams{
			Command: viper.GetString("command"),
			XmpPath: path,
			OnlyNew: viper.GetBool("new"),
			DryRun:  viper.GetBool("dry-run"),
		}
		err = xmp.Sync(params, viper.GetString("out"))
		if err != nil {
			return err
		}
	// raw
	case caseInsensitiveContains(viper.GetStringSlice("extension"), ext):
		fmt.Println("Syncing raw file with extension", ext, ":", path)
		raw, err := linkedimage.FindRaw(path, inDir, outDir)
		if err != nil {
			return err
		}
		params := darktable.ExportParams{
			Command: viper.GetString("command"),
			RawPath: path,
			OnlyNew: viper.GetBool("new"),
			DryRun:  viper.GetBool("dry-run"),
		}
		err = raw.Sync(params, viper.GetString("out"))
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
