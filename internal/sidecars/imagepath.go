package sidecars

import (
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

// ImagePath provides convenience methods to access various properties of an image file's path
// Works for raw files, jpg files, and xmp files that follow the convention <ParentDir>/<RelativeDir>/<ImageBase>_<VSequence>.<RawExt>.xmp
type ImagePath struct {
	fullPath string //Full relative or absolute path to image or sidecar file
	basePath string //Base file or directory (fullPath should match or be a sub path of basePath)
	baseDir  string //Base directory (derived from basePath)
}

func (i *ImagePath) GetFullPath() string {
	return i.fullPath
}

// GetBaseDir returns the directory where image files are found
func (i *ImagePath) GetBaseDir() string {
	if i.baseDir != "" {
		return i.baseDir
	}
	isDir, err := IsDir(i.basePath)
	if err != nil {
		log.Fatalf("Error getting relative dir for input path '%s': %v", i.basePath, err)
	}
	baseDir := i.basePath
	if !isDir {
		baseDir = filepath.Dir(baseDir)
	}
	i.baseDir = baseDir
	return i.baseDir
}

func (i *ImagePath) GetFullDir() string {
	return filepath.Dir(i.fullPath)
}

// GetBasename strips the directories and suffixes
//some/dir/subdir/DSC1234_01.ARW.xmp => DSC1234_01
func (i *ImagePath) GetBasename() string {
	basename := filepath.Base(i.fullPath)
	// Remove all extensions
	for filepath.Ext(basename) != "" {
		basename = strings.TrimSuffix(basename, filepath.Ext(basename))
	}
	return basename
}

// GetImageBase gets the base image name, without extensions or virtual copy sequences
// _DSc1234_01.ARW.xmp => _DSC1234
func (i *ImagePath) GetImageBase() string {
	basename := i.GetBasename()
	exp := regexp.MustCompile(`(.*)_\d\d$`)
	imageBase := exp.ReplaceAllString(basename, "${1}")
	return imageBase
}

// Get the base image name, without extensions or virtual copy sequences
// _DSc1234_01.ARW.xmp => _DSC1234
func (i *ImagePath) GetVSequence() string {
	basename := i.GetBasename()
	exp := regexp.MustCompile(`.*_(\d\d)$`)
	matches := exp.FindStringSubmatch(basename)
	var sequence string
	if len(matches) >= 1 {
		sequence = matches[1]
	}
	return sequence
}

// GetRelativeDir returns the directory of fullPath relative to baseDir
// E.g. GetRelativeDir("/mnt/some/dir/filename.txt", "/mnt") -> "some/dir"
func (i *ImagePath) GetRelativeDir() string {
	return filepath.Dir(i.GetRelativePath())
}

// GetRelativePath returns the value of fullPath relative to baseDir
// E.g. GetRelativePath("/mnt/some/dir/filename.txt", "/mnt") -> "some/dir/filename.txt"
func (i *ImagePath) GetRelativePath() string {
	baseDir := i.GetBaseDir()
	relativePath, err := filepath.Rel(baseDir, i.fullPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	// trace
	//fmt.Printf("trim '%s' off the front of '%s' => '%s'\n", baseDir, i.fullPath, relativePath)
	return relativePath
}
