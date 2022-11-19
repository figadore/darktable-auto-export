package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type options struct {
	inputFolder  string
	outputFolder string
	extension    string
	command      string
}

// Handle any complicated arg parsing here
func parseArgs() (*options, error) {
	opts := options{}
	opts.inputFolder = ""
	flag.StringVar(&opts.inputFolder, "i", "./", "Directory where raw files live")
	flag.StringVar(&opts.outputFolder, "o", "./", "Directory where jpgs go")
	flag.StringVar(&opts.extension, "e", ".ARW", "Extension of raw files")
	flag.StringVar(&opts.command, "c", "flatpak run --command=darktable-cli org.darktable.Darktable", "Darktable command or binary")
	flag.Parse()
	return &opts, nil
}

type Config struct {
	DeleteMissing string `yaml:"delete-missing"`
}

type exportParams struct {
	command    string
	rawPath    string
	xmpPath    string
	outputPath string
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

func main() {
	opts, err := parseArgs()
	if err != nil {
		log.Fatalf("Error parsing args: %v", err)
	}
	// Recurse through input directory
	raws := findFilesWithExt(opts.inputFolder, opts.extension)
	config := parseConfig()
	for _, raw := range raws {
		fmt.Println(raw)
		// Find adjacent xmp files
		xmps := findXmps(raw)
		basename := strings.TrimSuffix(filepath.Base(raw), filepath.Ext(raw))
		relativeDir := strings.TrimPrefix(filepath.Dir(raw), opts.inputFolder)
		outputPath := filepath.Join(opts.outputFolder, relativeDir, fmt.Sprintf("%s.jpg", basename))
		params := exportParams{
			command:    opts.command,
			rawPath:    raw,
			outputPath: outputPath,
		}
		if len(xmps) == 0 {
			fmt.Println("No xmp files found, applying default settings")
			export(params)
		} else {
			for _, xmp := range xmps {
				fmt.Println("  ", xmp)
				// Export the RAW file
				params.xmpPath = xmp
				jpgFilename := getJpgFilename(xmp, opts.extension)
				outputPath, err := filepath.Abs(filepath.Join(opts.outputFolder, relativeDir, jpgFilename))
				if err != nil {
					log.Fatalf("Error getting jpg path: %v", err)
				}
				params.outputPath = outputPath
				export(params)
			}
		}
	}
	// Delete jpgs for missing raws
	if config.DeleteMissing == "true" {
		fmt.Println("Deleting jpgs for missing raws")
		jpgs := findFilesWithExt(opts.outputFolder, ".jpg")
		deleteJpgs(findJpgsWithoutRaw(jpgs, opts.inputFolder, opts.outputFolder, opts.extension))
		fmt.Printf("Found %v jpgs", len(jpgs))
	} else {
		fmt.Printf("Not deleting jpgs for missing raws: %s", config.DeleteMissing)
	}
	// Look for xmp file(s) for the raw file
	// If no xmp file exists for a RAW...
	// Run darktable cli, setting export path to match structure of input dir
	//  darktable-cli [<input file or dir>] [<xmp file>] <output destination> [options] [--core <darktable options>]
	fmt.Println("\nComplete")
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

func export(params exportParams) {
	err := deleteJpgIfExists(params.outputPath)
	if err != nil {
		log.Fatalf("Error deleting jpg: %v", err)
	}
	//cmd := exec.Command("echo", params.rawPath, ":", params.xmpPath, "->", params.outputPath)
	args := strings.Fields(params.command)
	args = append(args, params.rawPath)
	if params.xmpPath != "" {
		args = append(args, params.xmpPath)
	}
	args = append(args, params.outputPath)
	remaining := args[1:]
	//cmd := exec.Command("echo", remaining...)
	fmt.Println(args)
	fmt.Println(len(args))
	cmd := exec.Command(args[0], remaining...)
	//cmd := exec.Command("echo", args...)
	stdout, err := cmd.CombinedOutput()
	fmt.Println("stdout", string(stdout))
	if err != nil {
		fmt.Println("error", err.Error())
		fmt.Println("err", err)
	}
}

// deleteJpgIfExists is what enables jpgs to be replaced rather than appended
// The darktable cli does not overwrite jpgs, it creates new ones every time it is run
func deleteJpgIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Found existing jpg at target path. removing %s\n", path)
		e := os.Remove(path)
		if e != nil {
			return e
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// file doesn't exist, nothing to delete
		return nil
	} else {
		return err
	}
	return nil
}
