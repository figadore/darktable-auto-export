package linkedimage

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO set up in-memory file system to make tests less brittle over time
func TestFindFilesWithExt(t *testing.T) {
	want := []string{"test/src/_DSC1234.ARW"}
	raws := FindFilesWithExt("./test/src", ".arw")
	if !reflect.DeepEqual(want, raws) {
		t.Fatalf(`Wanted %s, got %s`, want, raws)
	}
}

func TestRawGetJpgPath(t *testing.T) {
	var tests = []struct {
		rawPath string
		srcDir  string
		dstDir  string
		want    string
	}{
		{"/tests/src/_DSC1234.ARW", "/tests/src/", "/tests/dst", "/tests/dst/_DSC1234.jpg"},
		{"tests/src/subdir/_DSC1234.ARW", "tests/src/", "tests/dst", "tests/dst/subdir/_DSC1234.jpg"},
		{"tests/src/_DSC1234.arw", "tests/src", "tests/dst", "tests/dst/_DSC1234.jpg"},
		{"tests/src/_DSC1234.dng", "tests/src/", "tests/dst", "tests/dst/_DSC1234.jpg"},
		{"tests/src/_DSC1234_01.arw", "tests/src/", "tests/dst", "tests/dst/_DSC1234_01.jpg"},
		{"tests/src/_DSC1234_01", "tests/src/", "tests/dst", "tests/dst/_DSC1234_01.jpg"},
		{"tests/src/_DSC1234", "tests/src/", "tests/dst", "tests/dst/_DSC1234.jpg"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.rawPath)
		t.Run(testname, func(t *testing.T) {
			raw := NewRaw(ImagePath{fullPath: tt.rawPath, basePath: tt.srcDir})
			jpgPath := raw.GetJpgPath(tt.dstDir)
			if jpgPath != tt.want {
				t.Errorf("got %s, want %s", jpgPath, tt.want)
			}
		})
	}
}

func TestXmpGetJpgPath(t *testing.T) {
	var tests = []struct {
		xmpPath string
		srcDir  string
		dstDir  string
		want    string
	}{
		{"/tests/src/_DSC1234.ARW.xmp", "/tests/src/", "/tests/dst", "/tests/dst/_DSC1234.jpg"},
		{"tests/src/subdir/_DSC1234.ARW.xmp", "tests/src/", "tests/dst", "tests/dst/subdir/_DSC1234.jpg"},
		{"tests/src/_DSC1234.arw.xmp", "tests/src", "tests/dst", "tests/dst/_DSC1234.jpg"},
		{"tests/src/_DSC1234.dng.xmp", "tests/src/", "tests/dst", "tests/dst/_DSC1234.jpg"},
		{"tests/src/_DSC1234_01.arw.xmp", "tests/src/", "tests/dst", "tests/dst/_DSC1234_01.jpg"},
		{"tests/src/_DSC1234_01.xmp", "tests/src/", "tests/dst", "tests/dst/_DSC1234_01.jpg"},
		{"tests/src/_DSC1234.xmp", "tests/src/", "tests/dst", "tests/dst/_DSC1234.jpg"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.xmpPath)
		t.Run(testname, func(t *testing.T) {
			xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.srcDir})
			jpgPath := xmp.GetJpgPath(tt.dstDir)
			if jpgPath != tt.want {
				t.Errorf("got %s, want %s", jpgPath, tt.want)
			}
		})
	}
}

func TestXmpIsVirtualCopy(t *testing.T) {
	var tests = []struct {
		xmpPath string
		want    bool
	}{

		{"/some/dir/_DSC1234_01.arw.xmp", true},
		{"/some/dir/_DSC1234_01.xmp", true},
		{"/some/dir/_DSC1234_01.ARW.xmp", true},
		{"/some/dir/_DSC1234.dng.xmp", false},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.xmpPath)
		t.Run(testname, func(t *testing.T) {
			xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpPath})
			isVCopy := xmp.IsVirtualCopy()
			if isVCopy != tt.want {
				t.Errorf(`Wanted %v, got %v`, tt.want, isVCopy)
			}
		})
	}
}

func TestXmpGetRawExt(t *testing.T) {
	var tests = []struct {
		xmpPath string
		want    string
	}{

		{"/some/dir/_DSC1234_01.arw.xmp", ".arw"},
		{"/some/dir/_DSC1234_01.xmp", ""},
		{"/some/dir/_DSC1234_01.ARW.xmp", ".ARW"},
		{"/some/dir/_DSC1234.dng.xmp", ".dng"},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.xmpPath)
		t.Run(testname, func(t *testing.T) {
			xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpPath})
			rawExt := xmp.GetRawExt()
			if rawExt != tt.want {
				t.Errorf(`Wanted %v, got %v`, tt.want, rawExt)
			}
		})
	}
}

func TestLinkImages(t *testing.T) {
	rawPath1 := ImagePath{fullPath: "/src/_DSC0001.ARW", basePath: "/src"}
	xmpPath1 := ImagePath{fullPath: "/src/_DSC0001.ARW.xmp", basePath: "/src"}
	xmpPath2 := ImagePath{fullPath: "/src/_DSC0001_01.ARW.xmp", basePath: "/src"}
	jpgPath1 := ImagePath{fullPath: "/dst/_DSC0001.jpg", basePath: "/dst"}
	jpgPath2 := ImagePath{fullPath: "/dst/_DSC0001_01.jpg", basePath: "/dst"}

	var tests = []struct {
		name  string
		setup func() ([]Raw, []Xmp, []Jpg)
		raws  []ImagePath
		xmps  []ImagePath
		jpgs  []ImagePath
	}{
		{
			"jpg with no xmp or raw",
			func() ([]Raw, []Xmp, []Jpg) {
				jpg1 := NewJpg(jpgPath1)
				return []Raw{}, []Xmp{}, []Jpg{*jpg1}
			},
			[]ImagePath{},
			[]ImagePath{},
			[]ImagePath{jpgPath1},
		},
		{
			"xmp with jpg, no raw",
			func() ([]Raw, []Xmp, []Jpg) {
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				xmp1.LinkJpg(jpg1)
				return []Raw{}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
		},
		{
			"1 raw with 1 jpg, 1 xmp, linked by adding to raw",
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				// Link jpg and xmp directly to raw, but not to each other. Ensure cross link is created in the process
				raw1.AddXmp(xmp1)
				raw1.AddJpg(jpg1)
				return []Raw{*raw1}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
		},
		{
			"1 raw with 1 jpg, 1 xmp, xmp linked explicitly then added",
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				// Link jpg and xmp explicitly, then add to raw
				xmp1.LinkJpg(jpg1)
				raw1.AddXmp(xmp1)
				return []Raw{*raw1}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
		},
		{
			"1 raw with 1 jpg, 1 xmp, jpg linked explicitly then added",
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				// Link jpg and xmp explicitly, then add to raw
				jpg1.LinkXmp(xmp1)
				raw1.AddJpg(jpg1)
				return []Raw{*raw1}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
		},
		{
			"unlinked jpg and xmp",
			func() ([]Raw, []Xmp, []Jpg) {
				xmp1 := NewXmp(xmpPath2)
				jpg1 := NewJpg(jpgPath1)
				return []Raw{}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{},
			[]ImagePath{xmpPath2},
			[]ImagePath{jpgPath1},
		},
		{
			"raw with 2 xmp 2 jpg",
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				xmp2 := NewXmp(xmpPath2)
				jpg2 := NewJpg(jpgPath2)
				// try two different ways of linking
				raw1.AddXmp(xmp1)
				raw1.AddJpg(jpg1)
				raw1.AddXmp(xmp2)
				jpg2.LinkXmp(xmp2)
				jpg2.LinkRaw(raw1)
				return []Raw{*raw1}, []Xmp{*xmp1, *xmp2}, []Jpg{*jpg1, *jpg2}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1, xmpPath2},
			[]ImagePath{jpgPath1, jpgPath2},
		},
		// Other potential permutations
		// raw with 1 xmp no jpg
		// 1 unlinked raw, 1 xmp with no jpg
		// 1 unlinked raw, 1 linked raw with 1 xmp 1 jpg
		// raw with 1 jpg no xmp
		// raw with 2 xmp 3 jpg
		// xmp with raw and jpg
		// xmp with raw no jpg
		// xmp with jpg no raw
		// jpg with xmp no raw
		// jpg with raw no xmp
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			var raws []*Raw
			for _, raw := range tt.raws {
				raws = append(raws, NewRaw(raw))
			}
			var xmps []*Xmp
			for _, xmp := range tt.xmps {
				xmps = append(xmps, NewXmp(xmp))
			}
			var jpgs []*Jpg
			for _, jpg := range tt.jpgs {
				jpgs = append(jpgs, NewJpg(jpg))
			}
			linkImages(raws, xmps, jpgs)
			wantRaw, wantXmp, wantJpg := tt.setup()
			for i, want := range wantRaw {
				if want.String() != raws[i].String() {
					t.Errorf("Raw wanted \n%s\nbut got \n%s", want, raws[i])
					//} else {
					//	fmt.Printf("Wanted %s, and got %s\n", want, raws[i])
				}
			}
			for i, want := range wantXmp {
				if want.String() != xmps[i].String() {
					t.Errorf("Xmp wanted \n%s\nbut got \n%s", want, xmps[i])
				}
			}
			for i, want := range wantJpg {
				if want.String() != jpgs[i].String() {
					t.Errorf("Jpg wanted \n%s\nbut got \n%s", want, jpgs[i])
				}
			}
		})
	}
}

// Depends on TestLinkImages
func TestFindXmp(t *testing.T) {
	rawPath1 := ImagePath{fullPath: "test/src/_DSC1234.ARW", basePath: "test/src"}
	xmpPath1 := ImagePath{fullPath: "test/src/_DSC1234.ARW.xmp", basePath: "test/src"}
	jpgPath1 := ImagePath{fullPath: "test/dst/_DSC1234.jpg", basePath: "test/dst"}
	xmpPath2 := ImagePath{fullPath: "test/src/_DSC1234_01.ARW.xmp", basePath: "test/src"}
	jpgPath2 := ImagePath{fullPath: "test/dst/_DSC1234_01.jpg", basePath: "test/dst"}

	var tests = []struct {
		name    string
		xmpPath string
		setup   func() ([]Raw, []Xmp, []Jpg)
		raws    []ImagePath
		xmps    []ImagePath
		jpgs    []ImagePath
	}{
		{
			"xmp1 in tests/",
			xmpPath1.GetFullPath(),
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				raw1.AddXmp(xmp1)
				raw1.AddJpg(jpg1)
				return []Raw{*raw1}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1},
			[]ImagePath{jpgPath1},
		},
		{
			"xmp2 in tests/",
			xmpPath2.GetFullPath(),
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath2)
				jpg1 := NewJpg(jpgPath2)
				raw1.AddXmp(xmp1)
				raw1.AddJpg(jpg1)
				return []Raw{*raw1}, []Xmp{*xmp1}, []Jpg{*jpg1}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath2},
			[]ImagePath{jpgPath2},
		},
	}
	sourcesDir := "test/src"
	exportsDir := "test/dst"
	extensions := []string{".dng", ".arw"}
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			var raws []*Raw
			for _, raw := range tt.raws {
				raws = append(raws, NewRaw(raw))
			}
			var xmps []*Xmp
			for _, xmp := range tt.xmps {
				xmps = append(xmps, NewXmp(xmp))
			}
			var jpgs []*Jpg
			for _, jpg := range tt.jpgs {
				jpgs = append(jpgs, NewJpg(jpg))
			}
			linkImages(raws, xmps, jpgs)
			_, wantXmp, _ := tt.setup()
			xmp, err := FindXmp(tt.xmpPath, sourcesDir, exportsDir, extensions)
			if err != nil {
				t.Errorf("Failed to find xmp: %v", err)
			}
			if wantXmp[0].String() != xmp.String() {
				t.Errorf("Wanted %s but got %s", wantXmp[0], xmp)
			}
		})
	}
}

// Depends on TestLinkImages
func TestFindRaw(t *testing.T) {
	rawPath1 := ImagePath{fullPath: "test/src/_DSC1234.ARW", basePath: "test/src"}
	xmpPath1 := ImagePath{fullPath: "test/src/_DSC1234.ARW.xmp", basePath: "test/src"}
	jpgPath1 := ImagePath{fullPath: "test/dst/_DSC1234.jpg", basePath: "test/dst"}
	xmpPath2 := ImagePath{fullPath: "test/src/_DSC1234_01.ARW.xmp", basePath: "test/src"}
	jpgPath2 := ImagePath{fullPath: "test/dst/_DSC1234_01.jpg", basePath: "test/dst"}
	jpgPath3 := ImagePath{fullPath: "test/dst/_DSC1234_02.jpg", basePath: "test/dst"}

	var tests = []struct {
		name    string
		rawPath string
		setup   func() ([]Raw, []Xmp, []Jpg)
		raws    []ImagePath
		xmps    []ImagePath
		jpgs    []ImagePath
	}{
		{
			"raw from test/src",
			rawPath1.GetFullPath(),
			func() ([]Raw, []Xmp, []Jpg) {
				raw1 := NewRaw(rawPath1)
				xmp1 := NewXmp(xmpPath1)
				jpg1 := NewJpg(jpgPath1)
				xmp2 := NewXmp(xmpPath2)
				jpg2 := NewJpg(jpgPath2)
				jpg3 := NewJpg(jpgPath3)
				raw1.AddXmp(xmp1)
				raw1.AddJpg(jpg1)
				raw1.AddXmp(xmp2)
				raw1.AddJpg(jpg2)
				raw1.AddJpg(jpg3)
				return []Raw{*raw1}, []Xmp{*xmp1, *xmp2}, []Jpg{*jpg1, *jpg2, *jpg3}
			},
			[]ImagePath{rawPath1},
			[]ImagePath{xmpPath1, xmpPath2},
			[]ImagePath{jpgPath1, jpgPath2},
		},
	}
	sourcesDir := "test/src"
	exportsDir := "test/dst"
	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.name)
		t.Run(testname, func(t *testing.T) {
			var raws []*Raw
			for _, raw := range tt.raws {
				raws = append(raws, NewRaw(raw))
			}
			var xmps []*Xmp
			for _, xmp := range tt.xmps {
				xmps = append(xmps, NewXmp(xmp))
			}
			var jpgs []*Jpg
			for _, jpg := range tt.jpgs {
				jpgs = append(jpgs, NewJpg(jpg))
			}
			linkImages(raws, xmps, jpgs)
			wantRaw, _, _ := tt.setup()
			raw, err := FindRaw(tt.rawPath, sourcesDir, exportsDir)
			if err != nil {
				t.Errorf("Failed to find raw: %v", err)
			}
			if wantRaw[0].String() != raw.String() {
				t.Errorf("Wanted \n%s but got \n%s", wantRaw[0], raw)
			}
		})
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
		testname := fmt.Sprintf("%s==%s", tt.xmpPath, tt.rawPath)
		t.Run(testname, func(t *testing.T) {
			xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpPath})
			raw := NewRaw(ImagePath{fullPath: tt.rawPath, basePath: tt.rawPath})
			matches := xmpMatchesRaw(xmp, raw)
			if tt.want != matches {
				t.Errorf(`Wanted %v, got %v`, tt.want, matches)
			}
		})
	}
}

func TestJpgMatchesXmp(t *testing.T) {
	var tests = []struct {
		jpgPath    string
		jpgBaseDir string
		xmpPath    string
		xmpBaseDir string
		want       bool
	}{

		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01.arw.xmp", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01.ARW.xmp", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01.xmp", "/some/src", true},
		{"/some/dst/dir/_DSC1234.jpg", "/some/dst", "/some/src/dir/_DSC1234.ARW.xmp", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_02.xmp", "/some/src", false},
		{"some/dst/dir/_DSC1234_01.jpg", "some/dst", "some/src/dir/_DSC1234_02.xmp", "some/src", false},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s==%s", tt.xmpPath, tt.jpgPath)
		t.Run(testname, func(t *testing.T) {
			jpg := NewJpg(ImagePath{fullPath: tt.jpgPath, basePath: tt.jpgBaseDir})
			xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpBaseDir})
			matches := jpgMatchesXmp(jpg, xmp)
			if tt.want != matches {
				t.Errorf(`Wanted %v, got %v`, tt.want, matches)
			}
		})
	}
}

func TestJpgMatchesRaw(t *testing.T) {
	var tests = []struct {
		jpgPath    string
		jpgBaseDir string
		rawPath    string
		rawBaseDir string
		want       bool
	}{

		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01.arw", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01.ARW", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_01", "/some/src", true},
		{"/some/dst/dir/_DSC1234.jpg", "/some/dst", "/some/src/dir/_DSC1234.ARW", "/some/src", true},
		{"/some/dst/dir/_DSC1234_01.jpg", "/some/dst", "/some/src/dir/_DSC1234_02", "/some/src", false},
		{"/some/dst/dir/_DSC0018.jpg", "/some/dst/dir", "/some/src/dir/_DSC0018.ARW", "/some/src/dir", true},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s==%s", tt.rawPath, tt.jpgPath)
		t.Run(testname, func(t *testing.T) {
			fmt.Println("jpgBaseDir:", tt.jpgBaseDir)
			jpg := NewJpg(ImagePath{fullPath: tt.jpgPath, basePath: tt.jpgBaseDir})
			raw := NewRaw(ImagePath{fullPath: tt.rawPath, basePath: tt.rawBaseDir})
			matches := jpgMatchesRaw(jpg, raw)
			if tt.want != matches {
				t.Errorf(`Wanted %v, got %v`, tt.want, matches)
			}
		})
	}
}

// old
//func TestRawFindXmps(t *testing.T) {
//	want := []string{"test/src/_DSC1234.ARW.xmp", "test/src/_DSC1234_01.ARW.xmp"}
//	raws := FindFilesWithExt("./test/src", ".arw")
//	xmps := FindXmps(raws[0])
//	if !reflect.DeepEqual(want, xmps) {
//		t.Fatalf(`Wanted %s, got %s`, want, xmps)
//	}
//}
//
////func TestGetRelativeDir(t *testing.T) {
////	var tests = []struct {
////		fullPath  string
////		inputPath string
////		want      string
////	}{
////		{"./test/src/_DSC1234.ARW", "./test/src/_DSC1234.ARW.xmp", "."},
////		{"./test/src/_DSC1234.ARW", "./test/dst/_DSC4321.jpg", "../src"},
////		{"./the/path/filename.txt", ".", "the/path"},
////		{"/mnt/path/filename.txt", "/mnt", "path"},
////	}
////
////	for _, tt := range tests {
////		testname := fmt.Sprintf("%s:%s", tt.fullPath, tt.inputPath)
////		t.Run(testname, func(t *testing.T) {
////			relativeDir := GetRelativeDir(tt.fullPath, tt.inputPath)
////			if relativeDir != tt.want {
////				t.Errorf("got %s, want %s", relativeDir, tt.want)
////			}
////		})
////	}
////}
//
//func TestStripCommonDir(t *testing.T) {
//	var tests = []struct {
//		fullPath  string
//		inputPath string
//		want      string
//	}{
//		{"./test/src/_DSC1234.ARW", "./test/src/_DSC1234.ARW", "_DSC1234.ARW"},
//		{"./test/src/_DSC1234.ARW", "./test/src/", "_DSC1234.ARW"},
//		{"./test/dst/_DSC1234.jpg", "./test/dst/", "_DSC1234.jpg"},
//		//{"./test/src/_DSC1234.ARW", "./test/dst/", "src/_DSC1234.ARW"},
//		//{"./test/src/_DSC1234.ARW", "./test/dst/_DSC4321.jpg", "src/_DSC1234.ARW"},
//		{"./test/src/_DSC1234.ARW", "./test/", "src/_DSC1234.ARW"},
//	}
//
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s:%s", tt.fullPath, tt.inputPath)
//		t.Run(testname, func(t *testing.T) {
//			stripped := StripSharedDir(tt.fullPath, tt.inputPath)
//			if stripped != tt.want {
//				t.Errorf("got %s, want %s", stripped, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetJpgFilename(t *testing.T) {
//}
//
//func TestGetRawFilenameForJpg(t *testing.T) {
//	var tests = []struct {
//		jpgPath string
//		want    string
//	}{
//		{"tests/dst/_DSC1234.jpg", "_DSC1234.ARW"},
//		{"tests/dst/_DSC1234_123_01.jpg", "_DSC1234_123.ARW"},
//		{"tests/dst/_DSC1234_01.jpg", "_DSC1234.ARW"},
//	}
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s", tt.jpgPath)
//		t.Run(testname, func(t *testing.T) {
//			rawPath := GetRawFilenameForJpg(tt.jpgPath, ".ARW")
//			if rawPath != tt.want {
//				t.Errorf("got %s, want %s", rawPath, tt.want)
//			}
//		})
//	}
//}
//
//func TestGetXmpFilenameForJpg(t *testing.T) {
//	var tests = []struct {
//		jpgPath         string
//		want            string
//		wantVirtualCopy bool
//	}{
//		{"tests/dst/_DSC1234.jpg", "_DSC1234.ARW.xmp", false},
//		{"tests/dst/_DSC1234_123_01.jpg", "_DSC1234_123_01.ARW.xmp", true},
//		{"tests/dst/_DSC1234_123.jpg", "_DSC1234_123.ARW.xmp", false},
//		{"tests/dst/_DSC1234_01.jpg", "_DSC1234_01.ARW.xmp", true},
//	}
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s", tt.jpgPath)
//		t.Run(testname, func(t *testing.T) {
//			rawPath, isVirtualCopy := GetXmpFilenameForJpg(tt.jpgPath, ".ARW")
//			if rawPath != tt.want {
//				t.Errorf("got %s, want %s", rawPath, tt.want)
//			}
//			if isVirtualCopy != tt.wantVirtualCopy {
//				t.Errorf("for virtual copy check, got %v, want %v", isVirtualCopy, tt.wantVirtualCopy)
//			}
//		})
//	}
//}
//
//func TestFindJpgsWithoutRaw(t *testing.T) {
//	var tests = []struct {
//		rawExt []string
//		want   []string
//	}{
//		{[]string{".arw"}, []string{"test/dst/_DSC4321.jpg"}},
//		{[]string{".ARW"}, []string{"test/dst/_DSC4321.jpg"}},
//		{[]string{".ARW", ".dng"}, []string{"test/dst/_DSC4321.jpg"}},
//	}
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s", tt.rawExt)
//		t.Run(testname, func(t *testing.T) {
//			jpgs := FindFilesWithExt("./test/dst", ".jpg")
//			raws := FindFilesWithExt("./test/src", ".ARW")
//
//			//fmt.Println("jpgs:", jpgs, "raws:", raws)
//			jpgsToDelete := FindJpgsWithoutRaw(jpgs, raws, "test/src", "test/dst", tt.rawExt)
//			if !reflect.DeepEqual(tt.want, jpgsToDelete) {
//				t.Errorf(`Wanted %s, got %s`, tt.want, jpgsToDelete)
//			}
//		})
//	}
//}
//
//func TestFindJpgsWithoutXmp(t *testing.T) {
//	var tests = []struct {
//		rawExt []string
//		want   []string
//	}{
//		{[]string{".arw"}, []string{"test/dst/_DSC1234_02.jpg"}},
//		{[]string{".ARW"}, []string{"test/dst/_DSC1234_02.jpg"}},
//		{[]string{".ARW", ".dng"}, []string{"test/dst/_DSC1234_02.jpg"}},
//	}
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s", tt.rawExt)
//		t.Run(testname, func(t *testing.T) {
//			jpgs := FindFilesWithExt("./test/dst", ".jpg")
//			xmps := FindFilesWithExt("./test/src", ".xmp")
//			//raws := FindFilesWithExt("./test/src", ".ARW")
//			jpgsToDelete := FindJpgsWithoutXmp(jpgs, xmps, "test/src", "test/dst", tt.rawExt)
//			if !reflect.DeepEqual(tt.want, jpgsToDelete) {
//				t.Errorf(`Wanted %s, got %s`, tt.want, jpgsToDelete)
//			}
//		})
//	}
//}
//
//func TestGetRawPathForXmp(t *testing.T) {
//	var tests = []struct {
//		xmpPath string
//		want    string
//	}{
//
//		{"/some/dir/_DSC1234_01.arw.xmp", "/some/dir/_DSC1234.ARW"},
//		{"/some/dir/_DSC1234_01.xmp", "/some/dir/_DSC1234.ARW"},
//		{"/some/dir/_DSC1234_01.ARW.xmp", "/some/dir/_DSC1234.ARW"},
//	}
//	for _, tt := range tests {
//		testname := fmt.Sprintf("%s", tt.xmpPath)
//		t.Run(testname, func(t *testing.T) {
//			rawPath := getRawPathForXmp(tt.xmpPath, ".ARW")
//			if !reflect.DeepEqual(tt.want, rawPath) {
//				t.Errorf(`Wanted %s, got %s`, tt.want, rawPath)
//			}
//		})
//	}
//}
//
////func TestFindImages(t *testing.T) {
////	raws := FindImages("test/src", "test/dst", []string{".ARW", ".dng"})
////	// Print all raws
////	for _, raw := range raws {
////		fmt.Println(raw)
////		fmt.Println("xmps")
////		for _, xmp := range raw.Xmps {
////			fmt.Println("xmp:", xmp)
////			fmt.Println("jpg:", xmp.Jpg)
////		}
////		fmt.Println("jpgs")
////		for _, jpg := range raw.Jpgs {
////			fmt.Println("jpg:", jpg)
////			fmt.Println("xmp:", jpg.Xmp)
////		}
////	}
////}
//
//// No longer used, keeping in case needed in the future
////func TestXmpGetJpgPath(t *testing.T) {
////	var tests = []struct {
////		xmpPath    string
////		xmpBaseDir string
////		jpgBaseDir string
////		want       string
////	}{
////
////		{"/some/src/dir/_DSC1234_01.arw.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234_01.jpg"},
////		{"/some/src/dir/_DSC1234_01.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234_01.jpg"},
////		{"/some/src/dir/_DSC1234_01.ARW.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234_01.jpg"},
////		{"/some/src/dir/_DSC1234.dng.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234.jpg"},
////	}
////	for _, tt := range tests {
////		xmp := NewXmp(ImagePath{fullPath: tt.xmpPath, basePath: tt.xmpBaseDir})
////		jpgPath := xmp.GetJpgPath(tt.jpgBaseDir)
////		if jpgPath != tt.want {
////			t.Errorf(`Wanted %v, got %v`, tt.want, jpgPath)
////		}
////	}
////}
////
////func TestJpgGetXmpPath(t *testing.T) {
////	var tests = []struct {
////		want       string
////		xmpBaseDir string
////		jpgBaseDir string
////		jpgPath    string
////	}{
////
////		{"/some/src/dir/_DSC1234_01.ARW.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234_01.jpg"},
////		{"/some/src/dir/_DSC1234.ARW.xmp", "/some/src", "/some/dst", "/some/dst/dir/_DSC1234.jpg"},
////	}
////	for _, tt := range tests {
////		jpg := NewJpg(ImagePath{fullPath: tt.jpgPath, basePath: tt.jpgBaseDir})
////		xmpPath := jpg.GetXmpPath(tt.xmpBaseDir, ".ARW")
////		if xmpPath != tt.want {
////			t.Errorf(`Wanted %v, got %v`, tt.want, xmpPath)
////		}
////	}
////}
