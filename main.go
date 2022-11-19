package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func findJpgsWithoutRaw(jpgs []string, inputFolder, outputFolder, rawExtension string) []string {
	var jpgsToDelete []string
	for _, jpg := range jpgs {
		relativeDir := strings.TrimPrefix(filepath.Dir(jpg), outputFolder)
		rawFilenameLower := getRawFilename(jpg, strings.ToLower(rawExtension))
		rawFilenameUpper := getRawFilename(jpg, strings.ToUpper(rawExtension))
		rawPathLower := filepath.Join(inputFolder, relativeDir, rawFilenameLower)
		rawPathUpper := filepath.Join(inputFolder, relativeDir, rawFilenameUpper)
		// Check for the uppercase and lowercase version of the raw extension
		if _, err := os.Stat(rawPathLower); errors.Is(err, os.ErrNotExist) {
			if _, e := os.Stat(rawPathUpper); errors.Is(e, os.ErrNotExist) {
				jpgsToDelete = append(jpgsToDelete, jpg)
			}
		}
	}
	return jpgsToDelete
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

// _DSC1234_01.ARW.xmp -> _DSC1234_01.jpg
func getJpgFilename(xmpPath string, extension string) string {
	basename := strings.TrimSuffix(filepath.Base(xmpPath), filepath.Ext(xmpPath))
	// remove extra extension suffix
	// FIXME allow _DSC1234.xmp format as well (used by adobe and others)
	exp := regexp.MustCompile(fmt.Sprintf(`(?i)(.*)%s(.*)`, extension))
	jpgBasename := exp.ReplaceAllString(basename, "${1}${2}")
	return fmt.Sprintf("%s.jpg", jpgBasename)

}

// _DSC1234_01.jpg -> _DSC1234.ARW
func getRawFilename(jpgPath string, extension string) string {
	// Remove directory and extension
	basename := strings.TrimSuffix(filepath.Base(jpgPath), filepath.Ext(jpgPath))
	// remove sidecar duplicates suffix (e.g. _01) if it exists
	//exp := regexp.MustCompile(`(.*)(_\d\d)?`)
	exp := regexp.MustCompile(`(.*)_\d\d`)
	jpgBasename := exp.ReplaceAllString(basename, "${1}")
	return fmt.Sprintf("%s%s", jpgBasename, extension)
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

func findFilesWithExt(folder, extension string) []string {
	var raws []string
	err := filepath.WalkDir(folder, func(path string, info fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if strings.EqualFold(filepath.Ext(path), extension) {
			raws = append(raws, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return raws
}

// May be more efficient to enumerate once and deal with strings from then on
func findXmps(rawPath string) []string {
	var xmps []string
	base := strings.TrimSuffix(filepath.Base(rawPath), filepath.Ext(rawPath))
	ext := filepath.Ext(rawPath)
	dir := filepath.Dir(rawPath)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, er error) error {
		// Check each file adjacent to the raw file to see if it matches a sidecar xmp pattern
		if er != nil {
			fmt.Println(er)
		}
		// basename.xmp
		if strings.EqualFold(path, fmt.Sprintf("%s/%s.xmp", dir, base)) {
			xmps = append(xmps, path)
		}
		// basename.ext.xmp
		if strings.EqualFold(path, fmt.Sprintf("%s.xmp", rawPath)) {
			xmps = append(xmps, path)
		}
		// basename_XX.xmp
		if found, e := filepath.Match(fmt.Sprintf("%s/%s_[0-9][0-9].xmp", dir, base), path); e != nil {
			if found {
				xmps = append(xmps, path)
			}
		}
		// basename_XX.ext.xmp
		pattern := fmt.Sprintf("%s/%s_[0-9][0-9]%s.xmp", dir, base, ext)
		found, e := filepath.Match(pattern, path)
		if e != nil {
			return e
		}
		if found {
			xmps = append(xmps, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return xmps
}

func findRaw(jpgPath string) []string {
	var xmps []string
	base := strings.TrimSuffix(filepath.Base(jpgPath), filepath.Ext(jpgPath))
	ext := filepath.Ext(jpgPath)
	dir := filepath.Dir(jpgPath)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, er error) error {
		// Check each file adjacent to the raw file to see if it matches a sidecar xmp pattern
		if er != nil {
			fmt.Println(er)
		}
		// basename.xmp
		if strings.EqualFold(path, fmt.Sprintf("%s/%s.xmp", dir, base)) {
			xmps = append(xmps, path)
		}
		// basename.ext.xmp
		if strings.EqualFold(path, fmt.Sprintf("%s.xmp", jpgPath)) {
			xmps = append(xmps, path)
		}
		// basename_XX.xmp
		if found, e := filepath.Match(fmt.Sprintf("%s/%s_[0-9][0-9].xmp", dir, base), path); e != nil {
			if found {
				xmps = append(xmps, path)
			}
		}
		// basename_XX.ext.xmp
		pattern := fmt.Sprintf("%s/%s_[0-9][0-9]%s.xmp", dir, base, ext)
		found, e := filepath.Match(pattern, path)
		if e != nil {
			return e
		}
		if found {
			xmps = append(xmps, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return xmps
}
