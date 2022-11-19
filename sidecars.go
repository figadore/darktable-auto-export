package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
