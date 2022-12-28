package cmd

import (
	"fmt"

	"github.com/figadore/darktable-auto-export/internal/linkedimage"
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
	raws, xmps, _ := linkedimage.FindImages(inDir, outDir, extensions)

	// Trace
	// Print all raws
	for _, raw := range raws {
		fmt.Println(raw)
		//fmt.Println("xmps")
		//for _, xmp := range raw.Xmps {
		//	fmt.Println("xmp:", xmp)
		//	fmt.Println("jpg:", xmp.Jpg)
		//}
		//fmt.Println("jpgs")
		//for _, jpg := range raw.Jpgs {
		//	fmt.Println("jpg:", jpg)
		//	fmt.Println("xmp:", jpg.Xmp)
		//}
	}

	rawsToDelete := make(map[string]linkedimage.Raw)
	xmpsToDelete := make(map[string]linkedimage.Xmp)
	// for each jpg without a matching raw or xmp, aggregate the raw and/or xmp (making sure to get the xmps if deleting the raw)
	for i := range raws {
		// Assign raw to a variable here instead of iteration variables because loops copy the values
		raw := raws[i]
		if len(raw.Jpgs) == 0 {
			rawsToDelete[raw.GetPath()] = raws[i]
			// Clean up any orphan xmps
			for j := range raw.Xmps {
				xmp := raw.Xmps[j]
				xmpsToDelete[xmp.GetPath()] = *xmp
			}
		}
	}
	for i := range xmps {
		xmp := xmps[i]
		if xmp.Jpg == nil {
			xmpsToDelete[xmp.GetPath()] = xmp
		}
	}

	// list all raw and xmp files to delete. prompt for confirmation
	for k := range rawsToDelete {
		fmt.Println("Delete raw", k)
	}
	for k := range xmpsToDelete {
		fmt.Println("Delete xmp", k)
	}

	// if "move" selected at prompt, stage files for deletion in a folder. keep a log so it can be undone

	// if "yes" selected, or confirm after move, delete all enumerated raw and xmp files
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().StringP("in", "i", "./", "Directory or file of raw image(s)")
	cleanCmd.Flags().StringP("out", "o", "./", "Directory where jpgs exist")
	cleanCmd.Flags().StringSliceP("extension", "e", []string{".ARW"}, "Extension of raw files")
	cleanCmd.Flags().Bool("dry-run", false, "Show actions that would be performed, but don't do them")

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
