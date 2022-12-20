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

func FindJpgsWithoutRaw(jpgs []string, raws []string, inputFolder, outputFolder string, rawExtensions []string) []string {
	relativeJpgs := make([]string, len(jpgs))
	for i, jpg := range jpgs {
		relativeJpgs[i] = StripSharedDir(jpg, outputFolder)
	}
	relativeRaws := make([]string, len(raws))
	for i, raw := range raws {
		relativeRaws[i] = StripSharedDir(raw, inputFolder)
	}
	//fmt.Println("relativeJpgs:", relativeJpgs, "relativeRaws:", relativeRaws)
	var jpgsToDelete []string
	for i, jpg := range relativeJpgs {
		//fmt.Println("Looking for raw for", jpg)
		relativeDir := GetRelativeDir(jpg, outputFolder)
		found := false
		for _, rawExtension := range rawExtensions {
			// Check for uppercase and lowercase variations of extension
			rawFilenameLower := GetRawFilenameForJpg(jpg, strings.ToLower(rawExtension))
			rawFilenameUpper := GetRawFilenameForJpg(jpg, strings.ToUpper(rawExtension))
			rawPathLower := filepath.Join(inputFolder, relativeDir, rawFilenameLower)
			rawPathUpper := filepath.Join(inputFolder, relativeDir, rawFilenameUpper)
			// Check for the uppercase and lowercase version of the raw extension
			if caseInsensitiveContains(relativeRaws, rawPathUpper) || caseInsensitiveContains(relativeRaws, rawPathLower) {
				found = true
			}
		}
		if !found {
			jpgsToDelete = append(jpgsToDelete, jpgs[i])
		}
	}
	fmt.Println("Jpgs to delete (no matching raw):", jpgsToDelete)
	return jpgsToDelete
}

func caseInsensitiveContains(haystack []string, needle string) bool {
	//fmt.Println("Checking if", needle, "exists in", haystack)
	for _, v := range haystack {
		if strings.EqualFold(needle, v) {
			return true
		}
	}
	return false
}

// FindJpgsWithoutXmp scans for jpgs that have no corresponding raw/xmp pair
// This should only affect darktable duplicates, such as _DSC1234_01.ARw.xmp/_DSC1234_01.jpg
func FindJpgsWithoutXmp(jpgs []string, inputFolder, outputFolder string, rawExtensions []string) []string {
	var jpgsToDelete []string
	for _, jpg := range jpgs {
		relativeDir := GetRelativeDir(jpg, outputFolder)
		found := false
		for _, rawExtension := range rawExtensions {
			// Check for uppercase and lowercase variations of extension
			rawFilenameLower, isVirtualCopy := GetXmpFilenameForJpg(jpg, strings.ToLower(rawExtension))
			rawFilenameUpper, _ := GetXmpFilenameForJpg(jpg, strings.ToUpper(rawExtension))
			if !isVirtualCopy {
				// skip because this function only cares about virtual copies
				found = true
				break
			}
			rawPathLower := filepath.Join(inputFolder, relativeDir, rawFilenameLower)
			rawPathUpper := filepath.Join(inputFolder, relativeDir, rawFilenameUpper)
			// Check for the uppercase and lowercase version of the raw extension
			if _, err := os.Stat(rawPathLower); err == nil {
				found = true
			}
			if _, err := os.Stat(rawPathUpper); err == nil {
				found = true
			}
		}
		if !found {
			jpgsToDelete = append(jpgsToDelete, jpg)
		}
	}
	fmt.Println("Jpgs to delete (no matching xmp):", jpgsToDelete)
	return jpgsToDelete
}

func IsDir(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	// This returns an *os.FileInfo type
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	// IsDir is short for fileInfo.Mode().IsDir()
	if fileInfo.IsDir() {
		return true, nil
	} else {
		// not a directory
		return false, nil
	}
}

// GetRelativeDir returns the directory of fullPath relative to baseDir
// E.g. GetRelativeDir("/mnt/some/dir/filename.txt", "/mnt") -> "some/dir"
func GetRelativeDir(fullPath, baseDir string) string {
	isDir, err := IsDir(baseDir)
	if err != nil {
		log.Fatalf("Error getting relative dir for input path '%s': %v", baseDir, err)
	}

	if !isDir {
		baseDir = filepath.Dir(baseDir)
	}
	relativePath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	//for strings.HasPrefix(relativePath, "../") {
	//	relativePath = strings.TrimPrefix(relativePath, "../")
	//}
	//fmt.Printf("trim '%s' off the front of '%s' => '%s'\n", baseDir, fullPath, relativePath)
	return filepath.Dir(relativePath)
}

// StripSharedDir returns a path, stripping what it has in common with the second path provided
// E.g. StripSharedDir("/mnt/some/dir/filename.txt", "/mnt") -> "/some/dir/filename.txt"
func StripSharedDir(fullPath, baseDir string) string {
	relativeDir := GetRelativeDir(fullPath, baseDir)
	//relativePath, err := filepath.Rel(baseDir, fullPath)
	//if err != nil {
	//	log.Fatalf("%v", err)
	//}
	relativePath := filepath.Join(relativeDir, filepath.Base(fullPath))
	if relativePath == "." {
		relativePath = filepath.Base(fullPath)
	}
	// Trim any preceding '../'
	for strings.HasPrefix(relativePath, "../") {
		relativePath = strings.TrimPrefix(relativePath, "../")
	}
	//fmt.Printf("trim '%s' off the front of '%s' => '%s'\n", baseDir, fullPath, relativePath)
	return relativePath
}

// _DSC1234_01.ARW.xmp -> _DSC1234_01.jpg
// _DSC1234_01.xmp -> _DSC1234_01.jpg
func GetJpgFilename(xmpPath string, extensions []string) string {
	basename := strings.TrimSuffix(filepath.Base(xmpPath), filepath.Ext(xmpPath))
	jpgBasename := basename
	// remove optional extra extension suffix
	// allow _DSC1234.xmp format as well (used by adobe and others)
	for _, ext := range extensions {
		exp := regexp.MustCompile(fmt.Sprintf(`(?i)(.*)%s(.*)`, ext))
		jpgBasename = exp.ReplaceAllString(jpgBasename, "${1}${2}")
	}
	return fmt.Sprintf("%s.jpg", jpgBasename)

}

// GetRawFilenameForJpg transforms a jpg file name into the raw file name that should have been used to generate it
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

// GetXmpFilenameForJpg transforms a jpg file name into the xmp file name that should have been used to generate it
// _DSC1234_01.jpg -> _DSC1234_01.ARW.xmp (does not support _DSC1234_01.xmp format)
func GetXmpFilenameForJpg(jpgPath string, extension string) (string, bool) {
	// Remove directory and extension
	basename := strings.TrimSuffix(filepath.Base(jpgPath), filepath.Ext(jpgPath))
	exp := regexp.MustCompile(`(.*)_\d\d$`)
	isVirtualCopy := exp.Match([]byte(basename))
	xmpFilename := fmt.Sprintf("%s%s.xmp", basename, extension)
	return xmpFilename, isVirtualCopy
}

// /some/dir/_DSC1234_01.xmp -> /some/dir/_DSC1234.ARW
// /some/dir/_DSC1234_01.ARW.xmp -> /some/dir/_DSC1234.ARW
func getRawPathForXmp(xmpPath string, extension string) string {
	// Remove xmp extension
	basename := strings.TrimSuffix(xmpPath, filepath.Ext(xmpPath))
	// Remove sidecar duplicates suffix (e.g. _01) if it exists
	// Also strip out the raw extension if it exists
	exp := regexp.MustCompile(fmt.Sprintf(`(.*?)(?i)(?:_\d{2})?(?:%s)?`, extension))
	xmpBasename := exp.ReplaceAllString(basename, "${1}")
	return fmt.Sprintf("%s%s", xmpBasename, extension)
}

// FindRawPathForXmp locates the corresponding raw file on disk given an xmp
func FindRawPathForXmp(xmpPath string, extensions []string) (string, error) {
	for _, ext := range extensions {
		raw := getRawPathForXmp(xmpPath, strings.ToLower(ext))
		//fmt.Println("Given", xmpPath, "and", ext, "Looking for", raw)
		if _, err := os.Stat(raw); err == nil {
			return raw, nil
		}
		raw = getRawPathForXmp(xmpPath, strings.ToUpper(ext))
		if _, err := os.Stat(raw); err == nil {
			return raw, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Unable to find a matching raw file for %s", xmpPath))
}

func FindFilesWithExt(folder, extension string) []string {
	var raws []string
	err := filepath.WalkDir(folder, func(path string, info fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		// #recycle is for synology
		// FIXME make ignore paths configurable
		if strings.Contains(path, "#recycle") {
			return nil
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

// FindXmps gets any adjacent .xmp files given a raw file
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
