package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFindRaws(t *testing.T) {
	want := []string{"test/_DSC1234.ARW"}
	raws := findRaws("./test", ".arw")
	if !reflect.DeepEqual(want, raws) {
		t.Fatalf(`Wanted %s, got %s`, want, raws)
	}
}

func TestFindXmps(t *testing.T) {
	want := []string{"test/_DSC1234.ARW.xmp", "test/_DSC1234_01.ARW.xmp"}
	raws := findRaws("./test", ".arw")
	xmps := findXmps(raws[0])
	if !reflect.DeepEqual(want, xmps) {
		t.Fatalf(`Wanted %s, got %s`, want, xmps)
	}
}

func TestGetJpgName(t *testing.T) {
	var tests = []struct {
		rawPath string
		want    string
	}{
		{"tests/_DSC1234.ARW.xmp", "_DSC1234.jpg"},
		{"tests/_DSC1234.arw.xmp", "_DSC1234.jpg"},
		{"tests/_DSC1234_01.arw.xmp", "_DSC1234_01.jpg"},
		{"tests/_DSC1234.xmp", "_DSC1234.jpg"},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.rawPath)
		t.Run(testname, func(t *testing.T) {
			jpgPath := getJpgFilename(tt.rawPath, ".arw")
			if jpgPath != tt.want {
				t.Errorf("got %s, want %s", jpgPath, tt.want)
			}
		})
	}
}
