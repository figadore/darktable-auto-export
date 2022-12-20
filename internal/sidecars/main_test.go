package sidecars

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO set up in-memory file system to make tests less brittle over time
func TestFindRaws(t *testing.T) {
	want := []string{"test/src/_DSC1234.ARW"}
	raws := FindFilesWithExt("./test/src", ".arw")
	if !reflect.DeepEqual(want, raws) {
		t.Fatalf(`Wanted %s, got %s`, want, raws)
	}
}

func TestFindXmps(t *testing.T) {
	want := []string{"test/src/_DSC1234.ARW.xmp", "test/src/_DSC1234_01.ARW.xmp"}
	raws := FindFilesWithExt("./test/src", ".arw")
	xmps := FindXmps(raws[0])
	if !reflect.DeepEqual(want, xmps) {
		t.Fatalf(`Wanted %s, got %s`, want, xmps)
	}
}

func TestGetRelativeDir(t *testing.T) {
	var tests = []struct {
		fullPath  string
		inputPath string
		want      string
	}{
		{"./test/src/_DSC1234.ARW", "./test/src/_DSC1234.ARW.xmp", "."},
		{"./test/src/_DSC1234.ARW", "./test/dst/_DSC4321.jpg", "../src"},
		{"./the/path/filename.txt", ".", "the/path"},
		{"/mnt/path/filename.txt", "/mnt", "path"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.fullPath, tt.inputPath)
		t.Run(testname, func(t *testing.T) {
			relativeDir := GetRelativeDir(tt.fullPath, tt.inputPath)
			if relativeDir != tt.want {
				t.Errorf("got %s, want %s", relativeDir, tt.want)
			}
		})
	}
}

func TestStripCommonDir(t *testing.T) {
	var tests = []struct {
		fullPath  string
		inputPath string
		want      string
	}{
		{"./test/src/_DSC1234.ARW", "./test/src/_DSC1234.ARW", "_DSC1234.ARW"},
		{"./test/src/_DSC1234.ARW", "./test/src/", "_DSC1234.ARW"},
		{"./test/dst/_DSC1234.jpg", "./test/dst/", "_DSC1234.jpg"},
		//{"./test/src/_DSC1234.ARW", "./test/dst/", "src/_DSC1234.ARW"},
		//{"./test/src/_DSC1234.ARW", "./test/dst/_DSC4321.jpg", "src/_DSC1234.ARW"},
		{"./test/src/_DSC1234.ARW", "./test/", "src/_DSC1234.ARW"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.fullPath, tt.inputPath)
		t.Run(testname, func(t *testing.T) {
			stripped := StripSharedDir(tt.fullPath, tt.inputPath)
			if stripped != tt.want {
				t.Errorf("got %s, want %s", stripped, tt.want)
			}
		})
	}
}

func TestGetJpgFilename(t *testing.T) {
	var tests = []struct {
		rawPath string
		want    string
	}{
		{"/tests/src/_DSC1234.ARW.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234.ARW.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234.arw.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234.dng.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234_01.arw.xmp", "_DSC1234_01.jpg"},
		{"tests/src/_DSC1234_01.xmp", "_DSC1234_01.jpg"},
		{"tests/src/_DSC1234.xmp", "_DSC1234.jpg"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.rawPath)
		t.Run(testname, func(t *testing.T) {
			jpgPath := GetJpgFilename(tt.rawPath, []string{".DNG", ".arw"})
			if jpgPath != tt.want {
				t.Errorf("got %s, want %s", jpgPath, tt.want)
			}
		})
	}
}

func TestGetRawFilenameForJpg(t *testing.T) {
	var tests = []struct {
		jpgPath string
		want    string
	}{
		{"tests/dst/_DSC1234.jpg", "_DSC1234.ARW"},
		{"tests/dst/_DSC1234_123_01.jpg", "_DSC1234_123.ARW"},
		{"tests/dst/_DSC1234_01.jpg", "_DSC1234.ARW"},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.jpgPath)
		t.Run(testname, func(t *testing.T) {
			rawPath := GetRawFilenameForJpg(tt.jpgPath, ".ARW")
			if rawPath != tt.want {
				t.Errorf("got %s, want %s", rawPath, tt.want)
			}
		})
	}
}

func TestGetXmpFilenameForJpg(t *testing.T) {
	var tests = []struct {
		jpgPath         string
		want            string
		wantVirtualCopy bool
	}{
		{"tests/dst/_DSC1234.jpg", "_DSC1234.ARW.xmp", false},
		{"tests/dst/_DSC1234_123_01.jpg", "_DSC1234_123_01.ARW.xmp", true},
		{"tests/dst/_DSC1234_123.jpg", "_DSC1234_123.ARW.xmp", false},
		{"tests/dst/_DSC1234_01.jpg", "_DSC1234_01.ARW.xmp", true},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.jpgPath)
		t.Run(testname, func(t *testing.T) {
			rawPath, isVirtualCopy := GetXmpFilenameForJpg(tt.jpgPath, ".ARW")
			if rawPath != tt.want {
				t.Errorf("got %s, want %s", rawPath, tt.want)
			}
			if isVirtualCopy != tt.wantVirtualCopy {
				t.Errorf("for virtual copy check, got %v, want %v", isVirtualCopy, tt.wantVirtualCopy)
			}
		})
	}
}

func TestFindJpgsWithoutRaw(t *testing.T) {
	var tests = []struct {
		rawExt []string
		want   []string
	}{
		{[]string{".arw"}, []string{"test/dst/_DSC4321.jpg"}},
		{[]string{".ARW"}, []string{"test/dst/_DSC4321.jpg"}},
		{[]string{".ARW", ".dng"}, []string{"test/dst/_DSC4321.jpg"}},
	}
	for _, tt := range tests {
		jpgs := FindFilesWithExt("./test/dst", ".jpg")
		raws := FindFilesWithExt("./test/src", ".ARW")

		//fmt.Println("jpgs:", jpgs, "raws:", raws)
		jpgsToDelete := FindJpgsWithoutRaw(jpgs, raws, "test/src", "test/dst", tt.rawExt)
		if !reflect.DeepEqual(tt.want, jpgsToDelete) {
			t.Errorf(`Wanted %s, got %s`, tt.want, jpgsToDelete)
		}
	}
}

func TestFindJpgsWithoutXmp(t *testing.T) {
	var tests = []struct {
		rawExt []string
		want   []string
	}{
		{[]string{".arw"}, []string{"test/dst/_DSC1234_02.jpg"}},
		{[]string{".ARW"}, []string{"test/dst/_DSC1234_02.jpg"}},
		{[]string{".ARW", ".dng"}, []string{"test/dst/_DSC1234_02.jpg"}},
	}
	for _, tt := range tests {
		jpgs := FindFilesWithExt("./test/dst", ".jpg")
		//raws := FindFilesWithExt("./test/src", ".ARW")
		jpgsToDelete := FindJpgsWithoutXmp(jpgs, "test/src", "test/dst", tt.rawExt)
		if !reflect.DeepEqual(tt.want, jpgsToDelete) {
			t.Errorf(`Wanted %s, got %s`, tt.want, jpgsToDelete)
		}
	}
}

func TestGetRawPathForXmp(t *testing.T) {
	var tests = []struct {
		xmpPath string
		want    string
	}{

		{"/some/dir/_DSC1234_01.arw.xmp", "/some/dir/_DSC1234.ARW"},
		{"/some/dir/_DSC1234_01.xmp", "/some/dir/_DSC1234.ARW"},
		{"/some/dir/_DSC1234_01.ARW.xmp", "/some/dir/_DSC1234.ARW"},
	}
	for _, tt := range tests {
		rawPath := getRawPathForXmp(tt.xmpPath, ".ARW")
		if !reflect.DeepEqual(tt.want, rawPath) {
			t.Errorf(`Wanted %s, got %s`, tt.want, rawPath)
		}
	}
}
