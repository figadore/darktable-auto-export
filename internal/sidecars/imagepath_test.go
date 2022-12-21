package sidecars

import (
	"fmt"
	"testing"
)

func TestIsDir(t *testing.T) {
	var tests = []struct {
		path string
		want bool
	}{
		{"./test/src/", true},
		{"./test/src", true},
		{"./test/src/_DSC1234.ARW", false},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.path)
		t.Run(testname, func(t *testing.T) {
			isDir, err := IsDir(tt.path)
			if err != nil {
				t.Errorf("Error while checking if path was a directory: %s", err)
			}
			if isDir != tt.want {
				t.Errorf("got %v, want %v", isDir, tt.want)
			}
		})
	}
}

func TestImagePathBaseDir(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir",
			},
			"/some/dir",
		},
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
			},
			"/some/dir/subdir",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetBaseDir()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestImagePathBasename(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir",
			},
			"DSC1234_01",
		},
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
			},
			"DSC1234_01",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetBasename()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestImagePathImageBase(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir",
			},
			"DSC1234",
		},
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
			},
			"DSC1234",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetImageBase()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestImagePathVSequence(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir",
			},
			"01",
		},
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
				basePath: "/some/dir/subdir/DSC1234_01.ARW.xmp",
			},
			"01",
		},
		{
			ImagePath{
				fullPath: "/some/dir/subdir/DSC1234.ARW.xmp",
				basePath: "/some/dir/subdir/",
			},
			"",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetVSequence()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestImagePathRelativePath(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "./test/src/_DSC1234.ARW",
				basePath: "./test/src/_DSC1234.ARW",
			},
			"_DSC1234.ARW",
		},
		{
			ImagePath{
				fullPath: "./test/src/_DSC1234.ARW",
				basePath: "./test/src/",
			},
			"_DSC1234.ARW",
		},
		{
			ImagePath{
				fullPath: "./test/dst/_DSC1234.jpg",
				basePath: "./test/dst",
			},
			"_DSC1234.jpg",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetRelativePath()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestImagePathGetRelativeDir(t *testing.T) {
	var tests = []struct {
		path ImagePath
		want string
	}{
		{
			ImagePath{
				fullPath: "./test/src/_DSC1234.ARW",
				basePath: "./test/src/_DSC1234.ARW",
			},
			".",
		},
		{
			ImagePath{
				fullPath: "./test/src/_DSC1234.ARW",
				basePath: "./test/dst/_DSC4321.jpg",
			},
			"../src",
		},
		{
			ImagePath{
				fullPath: "./the/path/filename.txt",
				basePath: ".",
			},
			"the/path",
		},
		{
			ImagePath{
				fullPath: "/mnt/path/filename.txt",
				basePath: "/mnt",
			},
			"path",
		},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s:%s", tt.path.fullPath, tt.path.basePath)
		t.Run(testname, func(t *testing.T) {
			got := tt.path.GetRelativeDir()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
