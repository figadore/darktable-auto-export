package sidecars

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

func FindJpgsWithoutRaw(jpgs []string, inputFolder, outputFolder, rawExtension string) []string {
	var jpgsToDelete []string
	for _, jpg := range jpgs {
		relativeDir := strings.TrimPrefix(filepath.Dir(jpg), outputFolder)
		rawFilenameLower := GetRawFilenameForJpg(jpg, strings.ToLower(rawExtension))
		rawFilenameUpper := GetRawFilenameForJpg(jpg, strings.ToUpper(rawExtension))
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
func GetJpgFilename(xmpPath string, extension string) string {
	basename := strings.TrimSuffix(filepath.Base(xmpPath), filepath.Ext(xmpPath))
	// remove extra extension suffix
	// FIXME allow _DSC1234.xmp format as well (used by adobe and others)
	exp := regexp.MustCompile(fmt.Sprintf(`(?i)(.*)%s(.*)`, extension))
	jpgBasename := exp.ReplaceAllString(basename, "${1}${2}")
	return fmt.Sprintf("%s.jpg", jpgBasename)

}

// _DSC1234_01.jpg -> _DSC1234.ARW
func GetRawFilenameForJpg(jpgPath string, extension string) string {
	// Remove directory and extension
	basename := strings.TrimSuffix(filepath.Base(jpgPath), filepath.Ext(jpgPath))
	// remove sidecar duplicates suffix (e.g. _01) if it exists
	//exp := regexp.MustCompile(`(.*)(_\d\d)?`)
	exp := regexp.MustCompile(`(.*)_\d\d`)
	jpgBasename := exp.ReplaceAllString(basename, "${1}")
	return fmt.Sprintf("%s%s", jpgBasename, extension)
}

// /some/dir/_DSC1234_01.xmp -> /some/dir/_DSC1234.ARW
// /some/dir/_DSC1234_01.ARW.xmp -> /some/dir/_DSC1234.ARW
func GetRawPathForXmp(xmpPath string, extension string) string {
	// Remove xmp extension
	basename := strings.TrimSuffix(xmpPath, filepath.Ext(xmpPath))
	// Remove sidecar duplicates suffix (e.g. _01) if it exists
	// Also strip out the raw extension if it exists
	exp := regexp.MustCompile(fmt.Sprintf(`(.*?)(?:_\d{2})?(?:%s)?`, extension))
	xmpBasename := exp.ReplaceAllString(basename, "${1}")
	return fmt.Sprintf("%s%s", xmpBasename, extension)
}

func FindFilesWithExt(folder, extension string) []string {
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
func FindXmps(rawPath string) []string {
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

// DeleteJpgIfExists is what enables jpgs to be replaced rather than appended
// The darktable cli does not overwrite jpgs, it creates new ones every time it is run
func DeleteJpgIfExists(path string) error {
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
