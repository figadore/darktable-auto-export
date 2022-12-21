package cmd

import (
	"fmt"

	"github.com/figadore/darktable-auto-export/internal/sidecars"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove raws and xmps where jpg has been deleted",
	Long: `Remove raws and xmps where jpg has been deleted

Only run this when all raws have been exported, as any jpgs missing in the output directory will trigger a deletion of the corresponding raw and/or xmp`,
	Run: clean,
}

func clean(cmd *cobra.Command, args []string) {
	inDir := viper.GetString("in")
	outDir := viper.GetString("out")
	extensions := viper.GetStringSlice("extension")
	jpgs := sidecars.FindFilesWithExt(outDir, ".jpg")
	xmps := sidecars.FindFilesWithExt(inDir, ".xmp")
	var raws []string
	// For each defined extension, add to the list of raws
	for _, ext := range extensions {
		raws = append(raws, sidecars.FindFilesWithExt(inDir, ext)...)
	}
	fmt.Println(jpgs, xmps)
	// for each jpg without a matching raw or xmp, aggregate the raw and/or xmp (make sure to get the xmps if deleting the raw)
	sourcesToDelete := sidecars.FindSourcesWithoutJpg()

	// list all raw and xmp files to delete. prompt for confirmation
	for _, sourceFile := range sourcesToDelete {
		fmt.Println("Delete", sourceFile)
	}

	// if "move" selected at prompt, stage files for deletion in a folder. keep a log so it can be undone

	// if "yes" selected, or confirm after move, delete all enumerated raw and xmp files
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringP("in", "i", "./", "Directory or file of raw image(s)")
	cleanCmd.Flags().StringP("out", "o", "./", "Directory where jpgs exist")
	cleanCmd.Flags().StringSliceP("extension", "e", []string{".ARW"}, "Extension of raw files")

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
	cleanCmd.PreRun = func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cleanCmd.Flags())
	}
}
