package linkedimage

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/figadore/darktable-auto-export/internal/darktable"
)

type LinkedImage interface {
	GetPath() string
	String() string
}

type Raw struct {
	Path   ImagePath
	Xmps   map[string]*Xmp
	Jpgs   map[string]*Jpg
	srcDir string // Base directory where source files are found
	dstDir string // Base directory where exported image files are found
}

func (i *Raw) GetPath() string {
	return i.Path.fullPath
}

func NewRaw(path ImagePath) *Raw {
	raw := Raw{
		Path:   path,
		Xmps:   make(map[string]*Xmp),
		Jpgs:   make(map[string]*Jpg),
		srcDir: path.GetBaseDir(),
	}
	return &raw
}

func (raw Raw) String() string {
	s := fmt.Sprintf("%s", raw.GetPath())
	// Iterate map keys deterministically
	// Get keys, sort keys, then access using sorted keys
	xmpKeys := make([]string, 0)
	for k := range raw.Xmps {
		xmpKeys = append(xmpKeys, k)
	}
	sort.Strings(xmpKeys)
	for _, key := range xmpKeys {
		xmp := raw.Xmps[key]
		if xmp.Jpg != nil {
			s = fmt.Sprintf("%v\n  %v => %v", s, xmp.GetPath(), xmp.Jpg.GetPath())
		} else {
			s = fmt.Sprintf("%v\n  %v", s, xmp.GetPath())
		}
	}
	jpgKeys := make([]string, 0)
	for k := range raw.Jpgs {
		jpgKeys = append(jpgKeys, k)
	}
	sort.Strings(jpgKeys)
	for _, key := range jpgKeys {
		jpg := raw.Jpgs[key]
		if jpg.Xmp != nil {
			s = fmt.Sprintf("%v\n  %v => %v", s, jpg.GetPath(), jpg.Xmp.GetPath())
		} else {
			s = fmt.Sprintf("%v\n  %v", s, jpg.GetPath())
		}
	}
	return fmt.Sprintf("%v", s)
}

func (raw *Raw) AddXmp(xmp *Xmp) {
	// TODO validate pattern
	// only add if it doesn't already exist
	if _, ok := raw.Xmps[xmp.GetPath()]; !ok {
		raw.Xmps[xmp.GetPath()] = xmp
		if xmp.Jpg != nil {
			raw.AddJpg(xmp.Jpg)
		}
		xmp.Raw = raw
	}
	// If there's already a matching jpg for this raw, link it to the xmp
	if xmp.Jpg == nil && len(raw.Jpgs) > 0 {
		for k := range raw.Jpgs {
			if jpgMatchesXmp(raw.Jpgs[k], xmp) {
				xmp.LinkJpg(raw.Jpgs[k])
			}
		}
	}
}

func (raw *Raw) AddJpg(jpg *Jpg) {
	// TODO validate pattern
	raw.dstDir = jpg.Path.GetBaseDir()
	// only add if it doesn't already exist
	if _, ok := raw.Jpgs[jpg.GetPath()]; !ok {
		raw.Jpgs[jpg.GetPath()] = jpg
		if jpg.Xmp != nil {
			raw.AddXmp(jpg.Xmp)
		}
		jpg.Raw = raw
	}
	// If there's already a matching xmp for this raw, link it to the jpg
	if jpg.Xmp == nil && len(raw.Xmps) > 0 {
		for k := range raw.Xmps {
			if jpgMatchesXmp(jpg, raw.Xmps[k]) {
				jpg.LinkXmp(raw.Xmps[k])
			}
		}
	}
}

// GetJpgPath gets the jpg filename for a raw file
func (raw *Raw) GetJpgPath(jpgDir string) string {
	base := raw.Path.GetBasename()           //e.g. _DSC1234_01
	relativeDir := raw.Path.GetRelativeDir() //e.g. src
	jpgRelativePath := fmt.Sprintf("%s.jpg", filepath.Join(relativeDir, base))
	return filepath.Join(jpgDir, jpgRelativePath)
}

func (raw *Raw) GetRawExt() string {
	return filepath.Ext(raw.GetPath())
}

// Sync finds any related xmps and exports jpgs
// Internally, it also links the jpgs to the xmps and raws
func (raw *Raw) Sync(exportParams darktable.ExportParams, dstDir string) error {
	if len(raw.Xmps) > 0 {
		for _, xmp := range raw.Xmps {
			exportParams.XmpPath = xmp.GetPath()
			err := xmp.Sync(exportParams, dstDir)
			if err != nil {
				return err
			}
		}
	} else {
		exportParams.OutputPath = raw.GetJpgPath(dstDir)
		err := darktable.Export(exportParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (raw *Raw) StageForDeletion(dryRun bool) error {
	newPath := filepath.Join(raw.srcDir, "delete", raw.Path.GetRelativeDir(), filepath.Base(raw.GetPath()))
	if dryRun {
		fmt.Println("Move", raw.GetPath(), "to", newPath)
		return nil
	}
	err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.Rename(raw.GetPath(), newPath)
	raw.Path.fullPath = newPath
	return err
}

func (raw *Raw) Delete(dryRun bool) error {
	fmt.Println("Delete", raw.GetPath())
	if dryRun {
		return nil
	}
	err := os.Remove(raw.GetPath())
	if err != nil {
		return err
	}
	// Unlink
	for _, xmp := range raw.Xmps {
		xmp.Raw = nil
	}
	for _, jpg := range raw.Jpgs {
		jpg.Raw = nil
	}
	return nil
}

type Xmp struct {
	Path ImagePath
	Raw  *Raw
	Jpg  *Jpg
}

func (i *Xmp) GetPath() string {
	return i.Path.fullPath
}

func NewXmp(path ImagePath) *Xmp {
	xmp := Xmp{
		Path: path,
	}
	return &xmp
}

func (xmp Xmp) String() string {
	s := xmp.GetPath()
	if xmp.Raw != nil {
		s = fmt.Sprintf("%v <= %v", s, xmp.Raw.GetPath())
	}
	if xmp.Jpg != nil {
		s = fmt.Sprintf("%v => %v", s, xmp.Jpg.GetPath())
	}
	return s
}

func (xmp *Xmp) LinkRaw(raw *Raw) {
	// TODO validate pattern
	raw.AddXmp(xmp)
	xmp.Raw = raw
	if xmp.Jpg != nil {
		xmp.Jpg.LinkRaw(raw)
	}
}

func (xmp *Xmp) LinkJpg(jpg *Jpg) {
	// TODO validate pattern
	xmp.Jpg = jpg
	jpg.Xmp = xmp
	if xmp.Raw != nil {
		xmp.Raw.AddJpg(jpg)
	}
}

func (xmp *Xmp) IsVirtualCopy() bool {
	vSeq := xmp.Path.GetVSequence()
	return vSeq != ""
}

// GetRawExt gets the extension of the linked raw file
// Does not work for Adobe style xmps where raw extension is missing
func (xmp *Xmp) GetRawExt() string {
	if xmp.Raw != nil {
		return xmp.Raw.GetRawExt()
	}
	exp := regexp.MustCompile(`^.+(\.[^.]+)\.[^.]+$`)
	matches := exp.FindStringSubmatch(xmp.GetPath())
	var rawExt string
	if len(matches) >= 1 {
		rawExt = matches[1]
	}
	return rawExt
}

// Gets file name without directory or extension(s)
func (xmp *Xmp) GetBasename() string {
	basename := filepath.Base(xmp.GetPath())
	for filepath.Ext(basename) != "" {
		basename = strings.TrimSuffix(basename, filepath.Ext(basename))
	}
	return basename
}

// GetJpgPath gets the jpg filename for an xmp file
// This implementation assumes the only thing after the first "." is 'xmp' or '<raw-ext>.xmp'
func (xmp *Xmp) GetJpgPath(jpgDir string) string {
	base := xmp.Path.GetBasename()           //e.g. _DSC1234_01
	relativeDir := xmp.Path.GetRelativeDir() //e.g. src
	jpgRelativePath := fmt.Sprintf("%s.jpg", filepath.Join(relativeDir, base))
	return filepath.Join(jpgDir, jpgRelativePath)
}

// This implementation requires the list of extensions from viper
// _DSC1234_01.ARW.xmp -> _DSC1234_01.jpg
// _DSC1234_01.xmp -> _DSC1234_01.jpg
//func (xmp *Xmp) GetJpgPath() string {
//	al
//	basename := strings.TrimSuffix(filepath.Base(xmp.GetPath()), filepath.Ext(xmp.GetPath()))
//	jpgBasename := basename
//	// remove optional extra extension suffix
//	// allow _DSC1234.xmp format as well (used by adobe and others)
//	for _, ext := range extensions {
//		exp := regexp.MustCompile(fmt.Sprintf(`(?i)(.*)%s(.*)`, ext))
//		jpgBasename = exp.ReplaceAllString(jpgBasename, "${1}${2}")
//	}
//	return fmt.Sprintf("%s.jpg", jpgBasename)
//}

// Sync finds any relate raw and exports jpgs
// Internally, it also links the jpgs to the xmp and raw
func (xmp *Xmp) Sync(exportParams darktable.ExportParams, dstDir string) error {
	exportParams.OutputPath = xmp.GetJpgPath(dstDir)
	exportParams.RawPath = xmp.Raw.GetPath()
	err := darktable.Export(exportParams)
	if err != nil {
		return err
	}
	return nil
}

func (xmp *Xmp) StageForDeletion(dryRun bool) error {
	newPath := filepath.Join(xmp.Path.GetBaseDir(), "delete", xmp.Path.GetRelativeDir(), filepath.Base(xmp.GetPath()))
	if dryRun {
		fmt.Println("Move", xmp.GetPath(), "to", newPath)
		return nil
	}
	err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.Rename(xmp.GetPath(), newPath)
	xmp.Path.fullPath = newPath
	return err
}

func (xmp *Xmp) Delete(dryRun bool) error {
	fmt.Println("Delete", xmp.GetPath())
	if dryRun {
		return nil
	}
	err := os.Remove(xmp.GetPath())
	if err != nil {
		return err
	}
	// Unlink
	if xmp.Jpg != nil {
		xmp.Jpg.Xmp = nil
	}
	if xmp.Raw != nil {
		raw := xmp.Raw
		if _, ok := raw.Xmps[xmp.GetPath()]; ok {
			delete(raw.Xmps, xmp.GetPath())
		}
	}
	return nil
}

type Jpg struct {
	Path ImagePath
	Raw  *Raw
	Xmp  *Xmp
}

func (i *Jpg) GetPath() string {
	return i.Path.fullPath
}

func NewJpg(path ImagePath) *Jpg {
	jpg := Jpg{
		Path: path,
	}
	return &jpg
}

func (jpg *Jpg) LinkRaw(raw *Raw) {
	// TODO validate pattern
	raw.AddJpg(jpg)
	jpg.Raw = raw
	if jpg.Xmp != nil {
		jpg.Xmp.LinkJpg(jpg)
	}
}

func (jpg *Jpg) LinkXmp(xmp *Xmp) {
	// TODO validate pattern
	xmp.Jpg = jpg
	jpg.Xmp = xmp
	if jpg.Raw != nil {
		jpg.Raw.AddJpg(jpg)
	}
}

func (jpg *Jpg) IsVirtualCopy() bool {
	vSeq := jpg.Path.GetVSequence()
	return vSeq != ""
}

// No longer needed, keeping bits that may be useful in the future
//func (jpg *Jpg) GetXmpPath(xmpDir string, rawExt string) string {
//	// Check if there is a linked xmp
//	// Check raw for xmps that match
//	// Contruct an xmp filename (Only works for _DSc1234.<raw-extension>.xmp format)
//	base := jpg.Path.GetBasename()           //e.g. _DSC1234_01
//	relativeDir := jpg.Path.GetRelativeDir() //e.g. src
//	xmpRelativePath := fmt.Sprintf("%s%s.xmp", filepath.Join(relativeDir, base), rawExt)
//	return filepath.Join(xmpDir, xmpRelativePath)
//}

func (jpg Jpg) String() string {
	s := jpg.GetPath()
	if jpg.Xmp != nil {
		s = fmt.Sprintf("%v <= %v", s, jpg.Xmp.GetPath())
	}
	if jpg.Raw != nil {
		s = fmt.Sprintf("%v <= %v", s, jpg.Raw.GetPath())
	}
	return s
}

func (jpg *Jpg) Delete(dryRun bool) error {
	fmt.Println("Delete", jpg.GetPath())
	if dryRun {
		return nil
	}
	err := os.Remove(jpg.GetPath())
	if err != nil {
		return err
	}
	// Unlink
	if jpg.Xmp != nil {
		jpg.Xmp.Jpg = nil
	}
	if jpg.Raw != nil {
		raw := jpg.Raw
		if _, ok := raw.Jpgs[jpg.GetPath()]; ok {
			delete(raw.Jpgs, jpg.GetPath())
		}
	}
	return nil
}

// FindFilesWithExt recursively scans a directory for files with the specified extension
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

// List all raws, xmps, and jpgs found in the sources and exports dir
// Each returned object includes any linked objects that were detected
func FindImages(sourcesDir, exportsDir string, extensions []string) ([]*Raw, []*Xmp, []*Jpg) {
	var raws []*Raw
	var rawPaths []string
	// Find all files with any of the given extensions
	for _, ext := range extensions {
		rawPaths = append(rawPaths, FindFilesWithExt(sourcesDir, ext)...)
	}
	xmpPaths := FindFilesWithExt(sourcesDir, ".xmp")
	jpgPaths := FindFilesWithExt(exportsDir, ".jpg")
	// Create a new Raw object for each found path
	for _, rawPath := range rawPaths {
		raw := NewRaw(ImagePath{fullPath: rawPath, basePath: sourcesDir})
		raws = append(raws, raw)
	}
	var xmps []*Xmp
	for _, xmpPath := range xmpPaths {
		xmp := NewXmp(ImagePath{fullPath: xmpPath, basePath: sourcesDir})
		xmps = append(xmps, xmp)
	}
	var jpgs []*Jpg
	for _, jpgPath := range jpgPaths {
		jpg := NewJpg(ImagePath{fullPath: jpgPath, basePath: exportsDir})
		jpgs = append(jpgs, jpg)
	}
	linkImages(raws, xmps, jpgs)
	return raws, xmps, jpgs
}

// FindXmp looks for an xmp file at the specified path
// The returned object includes any linked objects that were detected
func FindXmp(path, sourcesDir, exportsDir string, extensions []string) (*Xmp, error) {
	xmp := NewXmp(ImagePath{fullPath: path, basePath: sourcesDir})
	if !xmp.Path.Exists() {
		return nil, fmt.Errorf("Unable to find xmp at '%x'", path)
	}

	rawDir := xmp.Path.GetFullDir()
	var rawPaths []string
	for _, ext := range extensions {
		rawPaths = append(rawPaths, FindFilesWithExt(rawDir, ext)...)
	}
	var raws []*Raw
	// TODO search for limited number of raws like GetJpgPath ?
	for _, rawPath := range rawPaths {
		raw := NewRaw(ImagePath{fullPath: rawPath, basePath: sourcesDir})
		raws = append(raws, raw)
	}
	jpgDir := filepath.Join(exportsDir, xmp.Path.GetRelativeDir())
	jpg := NewJpg(ImagePath{fullPath: xmp.GetJpgPath(jpgDir), basePath: exportsDir})
	var jpgs []*Jpg
	if jpg.Path.Exists() {
		jpg := NewJpg(ImagePath{fullPath: jpg.GetPath(), basePath: exportsDir})
		jpgs = []*Jpg{jpg}
	}
	xmps := []*Xmp{xmp}
	linkImages(raws, xmps, jpgs)
	return xmp, nil
}

// FindRaw looks for a raw file at the specified path
// The returned object includes any linked objects that were detected
func FindRaw(path, sourcesDir, exportsDir string) (*Raw, error) {
	raw := NewRaw(ImagePath{fullPath: path, basePath: sourcesDir})
	if !raw.Path.Exists() {
		return nil, fmt.Errorf("Unable to find raw at '%x'", path)
	}

	// optimization compared to FindImages, look only in relativeDir
	xmpDir := raw.Path.GetFullDir()
	xmpPaths := FindFilesWithExt(xmpDir, ".xmp")
	var xmps []*Xmp
	for _, xmpPath := range xmpPaths {
		xmp := NewXmp(ImagePath{fullPath: xmpPath, basePath: sourcesDir})
		xmps = append(xmps, xmp)
	}
	jpgDir := filepath.Join(exportsDir, raw.Path.GetRelativeDir())
	var jpgs []*Jpg
	jpgPaths := FindFilesWithExt(jpgDir, ".jpg")
	for _, jpgPath := range jpgPaths {
		jpg := NewJpg(ImagePath{fullPath: jpgPath, basePath: exportsDir})
		jpgs = append(jpgs, jpg)
	}
	linkImages([]*Raw{raw}, xmps, jpgs)
	return raw, nil
}

// For each raw, find corresponding xmps and jpgs
// For each xmp, find corresponding jpgs and raws
// For each jpg, find corresponding xmps and raws
func linkImages(raws []*Raw, xmps []*Xmp, jpgs []*Jpg) {
	for i, raw := range raws {
		for j, xmp := range xmps {
			if xmpMatchesRaw(xmp, raw) {
				raws[i].AddXmp(xmps[j])
			}
		}
		for j, jpg := range jpgs {
			if jpgMatchesRaw(jpg, raw) {
				raws[i].AddJpg(jpgs[j])
			}
		}
	}
	for i, xmp := range xmps {
		for j, jpg := range jpgs {
			if jpgMatchesXmp(jpg, xmp) {
				xmps[i].LinkJpg(jpgs[j])
			}
		}
	}
	for i, jpg := range jpgs {
		for j, raw := range raws {
			if jpgMatchesRaw(jpg, raw) {
				jpgs[i].LinkRaw(raws[j])
			}
		}
	}
}

func jpgMatchesXmp(jpg *Jpg, xmp *Xmp) bool {
	jpgRelativePath := jpg.Path.GetRelativePath()
	base := xmp.Path.GetBasename()
	relativeDir := xmp.Path.GetRelativeDir()
	exp := regexp.MustCompile(fmt.Sprintf(`^(%s[\\\/])?%s\.jpg$`, relativeDir, base))
	if exp.Match([]byte(jpgRelativePath)) {
		return true
	}
	return false
}

func xmpMatchesRaw(xmp *Xmp, raw *Raw) bool {
	xmpPath := xmp.GetPath()
	base := raw.Path.GetBasename()
	ext := filepath.Ext(raw.GetPath())
	dir := raw.Path.GetFullDir()
	// basename.xmp
	exp := regexp.MustCompile(fmt.Sprintf(`^%s[\\\/]%s\.xmp$`, dir, base))
	if exp.Match([]byte(xmpPath)) {
		return true
	}
	// basename.ext.xmp
	exp = regexp.MustCompile(fmt.Sprintf(`^%s[\\\/]%s(?i)%s(?-i)\.xmp$`, dir, base, ext))
	if exp.Match([]byte(xmpPath)) {
		return true
	}
	// basename_XX.xmp
	exp = regexp.MustCompile(fmt.Sprintf(`^%s[\\\/]%s_\d\d\.xmp$`, dir, base))
	if exp.Match([]byte(xmpPath)) {
		return true
	}
	// basename_XX.ext.xmp
	exp = regexp.MustCompile(fmt.Sprintf(`^%s[\\\/]%s_\d\d(?i)%s(?-i)\.xmp$`, dir, base, ext))
	if exp.Match([]byte(xmpPath)) {
		return true
	}
	return false
}

func jpgMatchesRaw(jpg *Jpg, raw *Raw) bool {
	jpgPath := jpg.Path.GetRelativePath()
	base := raw.Path.GetBasename()
	ext := filepath.Ext(raw.GetPath())
	dir := raw.Path.GetRelativeDir()
	// basename.jpg
	exp := regexp.MustCompile(fmt.Sprintf(`^(%s[\\\/])?%s\.jpg$`, dir, base))
	if exp.Match([]byte(jpgPath)) {
		return true
	}
	// basename.ext.jpg
	exp = regexp.MustCompile(fmt.Sprintf(`^(%s[\\\/])?%s(?i)%s(?-i)\.jpg$`, dir, base, ext))
	if exp.Match([]byte(jpgPath)) {
		return true
	}
	// basename_XX.jpg
	exp = regexp.MustCompile(fmt.Sprintf(`^(%s[\\\/])?%s_\d\d\.jpg$`, dir, base))
	if exp.Match([]byte(jpgPath)) {
		return true
	}
	// basename_XX.ext.jpg
	exp = regexp.MustCompile(fmt.Sprintf(`^(%s[\\\/])?%s_\d\d(?i)%s(?-i)\.jpg$`, dir, base, ext))
	if exp.Match([]byte(jpgPath)) {
		return true
	}
	return false
}

// IsDir checks whether a path is a directory
// This implementation is naive, assuming anything not matching  "*.*" is a directory
func IsDir(path string) (bool, error) {
	exp := regexp.MustCompile(`^.+\..+$`)
	isFile := exp.Match([]byte(path))
	return !isFile, nil
}

// IsDir checks whether a path is a directory
// This implementation depends on i/o and requires files to actually exist
//func IsDir(path string) (bool, error) {
//	file, err := os.Open(path)
//	if err != nil {
//		return false, err
//	}
//	defer file.Close()
//	// This returns an *os.FileInfo type
//	fileInfo, err := file.Stat()
//	if err != nil {
//		return false, err
//	}
//
//	// IsDir is short for fileInfo.Mode().IsDir()
//	if fileInfo.IsDir() {
//		return true, nil
//	} else {
//		// not a directory
//		return false, nil
//	}
//}
