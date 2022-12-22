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
		xmps := FindFilesWithExt("./test/src", ".xmp")
		//raws := FindFilesWithExt("./test/src", ".ARW")
		jpgsToDelete := FindJpgsWithoutXmp(jpgs, xmps, "test/src", "test/dst", tt.rawExt)
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

//func TestFindImages(t *testing.T) {
//	raws := FindImages("test/src", "test/dst", []string{".ARW", ".dng"})
//	// Print all raws
//	for _, raw := range raws {
//		fmt.Println(raw)
//		fmt.Println("xmps")
//		for _, xmp := range raw.Xmps {
//			fmt.Println("xmp:", xmp)
//			fmt.Println("jpg:", xmp.Jpg)
//		}
//		fmt.Println("jpgs")
//		for _, jpg := range raw.Jpgs {
//			fmt.Println("jpg:", jpg)
//			fmt.Println("xmp:", jpg.Xmp)
//		}
//	}
//}

func TestLinkImages(t *testing.T) {
	rawPath1 := ImagePath{fullPath: "/src/_DSC0001.ARW", basePath: "/src"}
	xmpPath1 := ImagePath{fullPath: "/src/_DSC0001.ARW.xmp", basePath: "/src"}
	//xmpPath2 := ImagePath{fullPath: "/src/_DSC0001_01.ARW.xmp", basePath: "/src"}
	jpgPath1 := ImagePath{fullPath: "/dst/_DSC0001.jpg", basePath: "/dst"}
	//jpgPath2 := ImagePath{fullPath: "/dst/_DSC0001_01.jpg", basePath: "/dst"}

	linkedRaw1 := NewRaw(rawPath1)
	// TODO this is problematic, xmp and jpg not linked, make it more fail-safe
	//linkedRaw1.AddXmp(NewXmp(xmpPath1))
	//linkedRaw1.AddJpg(NewJpg(jpgPath1))
	linkedXmp1 := NewXmp(xmpPath1)
	linkedJpg1 := NewJpg(jpgPath1)
	linkedXmp1.AddJpg(linkedJpg1)
	linkedRaw1.AddXmp(linkedXmp1)

	var tests = []struct {
		raws    []ImagePath
		xmps    []ImagePath
		jpgs    []ImagePath
		wantRaw []Raw
		wantXmp []Xmp
		wantJpg []Jpg
	}{
		// raw with 1 jpg 1 xmp
		{
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
			[]Raw{*linkedRaw1},
			[]Xmp{*linkedXmp1},
			[]Jpg{*linkedJpg1},
		},
		// raw with no xmp or jpg
		// raw with 1 xmp no jpg
		// raw with 1 jpg no xmp
		// raw with 2 xmp 2 jpg
		// raw with 2 xmp 3 jpg
		// xmp with no raw or jpg
		// xmp with raw and jpg
		// xmp with raw no jpg
		// xmp with jpg no raw
		// jpg with no xmp or raw
		// jpg with xmp no raw
		// jpg with raw no xmp
	}
	for _, tt := range tests {
		var raws []Raw
		for _, raw := range tt.raws {
			raws = append(raws, *NewRaw(raw))
		}
		var xmps []Xmp
		for _, xmp := range tt.xmps {
			xmps = append(xmps, *NewXmp(xmp))
		}
		var jpgs []Jpg
		for _, jpg := range tt.jpgs {
			jpgs = append(jpgs, *NewJpg(jpg))
		}
		linkImages(raws, xmps, jpgs)
		// TODO sort these before compare?
		for i, want := range tt.wantRaw {
			if want.String() != raws[i].String() {
				t.Errorf(`Wanted %s, got %s`, want, raws[i])
			} else {
				t.Errorf(`Wanted %s, got %s`, want, raws[i])
			}
		}
		// TODO compare wanted xmps and jpgs too
	}
}

func TestXmpMatchesRaw(t *testing.T) {
	var tests = []struct {
		xmpPath string
		rawPath string
		want    bool
	}{

		{"/some/dir/_DSC1234_01.arw.xmp", "/some/dir/_DSC1234.ARW", true},
		{"/some/dir/_DSC1234_01.xmp", "/some/dir/_DSC1234.ARW", true},
		{"/some/dir/_DSC1234_01.ARW.xmp", "/some/dir/_DSC1234.ARW", true},
		{"/some/dir/_DSC1234.ARW.xmp", "/some/dir/_DSC1234.ARW", true},
		{"/some/dir/_DSC1234.xmp", "/some/dir/_DSC1234.ARW", true},
		{"/some/dir/_DSC1234_012.arw.xmp", "/some/dir/_DSC1234.ARW", false},
		{"/some/dir/_DSC0234.xmp", "/some/dir/_DSC1234.ARW", false},
	}
	for _, tt := range tests {
		xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpPath})
		raw := NewRaw(ImagePath{fullPath: tt.rawPath, basePath: tt.rawPath})
		matches := xmpMatchesRaw(*xmp, *raw)
		if tt.want != matches {
			t.Errorf(`Wanted %v, got %v`, tt.want, matches)
		}
	}
}
