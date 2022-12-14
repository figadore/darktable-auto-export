package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"log"

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
	//for _, raw := range raws {
	//    fmt.Println(raw)
	//    //fmt.Println("xmps")
	//    //for _, xmp := range raw.Xmps {
	//    //	fmt.Println("xmp:", xmp)
	//    //	fmt.Println("jpg:", xmp.Jpg)
	//    //}
	//    //fmt.Println("jpgs")
	//    //for _, jpg := range raw.Jpgs {
	//    //	fmt.Println("jpg:", jpg)
	//    //	fmt.Println("xmp:", jpg.Xmp)
	//    //}
	//}

	rawsToDelete := make(map[string]*linkedimage.Raw)
	xmpsToDelete := make(map[string]*linkedimage.Xmp)
	// for each jpg without a matching raw or xmp, aggregate the raw and/or xmp (making sure to get the xmps if deleting the raw)
	for _, raw := range raws {
		if len(raw.Jpgs) == 0 {
			rawsToDelete[raw.GetPath()] = raw
			// Clean up any orphan xmps
			for _, xmp := range raw.Xmps {
				xmpsToDelete[xmp.GetPath()] = xmp
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
	if len(xmpsToDelete) > 0 || len(rawsToDelete) > 0 {

		doStage := YesNoPrompt("Stage files listed above for deletion?", false)

		if doStage {
			fmt.Println("Staging raws")
			for k := range rawsToDelete {
				raw := rawsToDelete[k]
				fmt.Println("Staging", raw)
				err := raw.StageForDeletion(viper.GetBool("dry-run"))
				if err != nil {
					log.Fatalf("Error staging %s for deletion: %v", raw.GetPath(), err)
				}
			}
			fmt.Println("Staging xmps")
			for k := range xmpsToDelete {
				xmp := xmpsToDelete[k]
				fmt.Println("Staging", xmp)
				err := xmp.StageForDeletion(viper.GetBool("dry-run"))
				fmt.Println("xmp after staging", xmp, ":", xmp.GetPath())
				if err != nil {
					log.Fatalf("Error staging %s for deletion: %v", xmp.GetPath(), err)
				}
			}
		}

		// if "move" selected at prompt, stage files for deletion in a folder. keep a log so it can be undone

		// if "yes" selected, or confirm after move, delete all enumerated raw and xmp files
		doDelete := YesNoPrompt("Delete the source files listed above?", false)
		if doDelete {
			fmt.Println("Deleting raws")
			for k := range rawsToDelete {
				raw := rawsToDelete[k]
				fmt.Println("Deleting", raw)
				err := raw.Delete(viper.GetBool("dry-run"))
				if err != nil {
					log.Fatalf("Error deleting %s: %v", raw.GetPath(), err)
				}
			}
			fmt.Println("Deleting xmps")
			for k := range xmpsToDelete {
				xmp := xmpsToDelete[k]
				fmt.Println("Deleting", xmp)
				err := xmp.Delete(viper.GetBool("dry-run"))
				if err != nil {
					log.Fatalf("Error deleting %s: %v", xmp.GetPath(), err)
				}
			}
		}
		fmt.Println("doDelete:", doDelete)
	} else {
		fmt.Println("No candidate source files to delete")
	}
}

// YesNoPrompt asks yes/no questions using the label.
func YesNoPrompt(label string, defaultChoice bool) bool {
	choices := "Y/n"
	if !defaultChoice {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return defaultChoice
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
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
