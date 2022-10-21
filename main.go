package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
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
	flag.StringVar(&opts.extension, "e", ".arw", "Extension of raw files")
	flag.StringVar(&opts.command, "c", "flatpak run --command='darktable-cli' org.darktable.Darktable", "Darktable command or binary")
	flag.Parse()
	//args := flag.Args()
	//if len(args) < 1 {
	//	return &options{}, fmt.Errorf("expecting at least 1 arg (secret_name), found %d", len(args))
	//}
	return &opts, nil
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		log.Fatalf("Error parsing args: %v", err)
	}
	// Recurse through input directory
	raws := findRaws(opts.inputFolder, opts.extension)
	for _, raw := range raws {
		fmt.Println(raw)
		// Find adjacent xmp files
		xmps := findXmps(raw)
		for _, xmp := range xmps {
			fmt.Println("  ", xmp)
		}
	}
	// Look for xmp file(s) for the raw file
	// If no xmp file exists for a RAW...
	// Run darktable cli, setting export path to match structure of input dir
	//  darktable-cli [<input file or dir>] [<xmp file>] <output destination> [options] [--core <darktable options>]
	fmt.Println("\nComplete")
}

func findRaws(folder, extension string) []string {
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
