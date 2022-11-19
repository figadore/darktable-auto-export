package main

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO set up in-memory file system to make tests less brittle over time
func TestFindRaws(t *testing.T) {
	want := []string{"test/src/_DSC1234.ARW"}
	raws := findFilesWithExt("./test/src", ".arw")
	if !reflect.DeepEqual(want, raws) {
		t.Fatalf(`Wanted %s, got %s`, want, raws)
	}
}

func TestFindXmps(t *testing.T) {
	want := []string{"test/src/_DSC1234.ARW.xmp", "test/src/_DSC1234_01.ARW.xmp"}
	raws := findFilesWithExt("./test/src", ".arw")
	xmps := findXmps(raws[0])
	if !reflect.DeepEqual(want, xmps) {
		t.Fatalf(`Wanted %s, got %s`, want, xmps)
	}
}

func TestGetJpgFilename(t *testing.T) {
	var tests = []struct {
		rawPath string
		want    string
	}{
		{"tests/src/_DSC1234.ARW.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234.arw.xmp", "_DSC1234.jpg"},
		{"tests/src/_DSC1234_01.arw.xmp", "_DSC1234_01.jpg"},
		{"tests/src/_DSC1234.xmp", "_DSC1234.jpg"},
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

func TestGetRawFilename(t *testing.T) {
	var tests = []struct {
		jpgPath string
		want    string
	}{
		{"tests/dst/_DSC1234.jpg", "_DSC1234.ARW"},
		{"tests/dst/_DSC1234_01.jpg", "_DSC1234.ARW"},
	}
	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.jpgPath)
		t.Run(testname, func(t *testing.T) {
			rawPath := getRawFilename(tt.jpgPath, ".ARW")
			if rawPath != tt.want {
				t.Errorf("got %s, want %s", rawPath, tt.want)
			}
		})
	}
}

func TestFindJpgsWithoutRaw(t *testing.T) {
	var tests = []struct {
		rawExt string
		want   []string
	}{
		{".arw", []string{"test/dst/_DSC4321.jpg"}},
		{".ARW", []string{"test/dst/_DSC4321.jpg"}},
	}
	for _, tt := range tests {
		jpgs := findFilesWithExt("./test/dst", ".jpg")
		jpgsToDelete := findJpgsWithoutRaw(jpgs, "test/src", "test/dst", tt.rawExt)
		if !reflect.DeepEqual(tt.want, jpgsToDelete) {
			t.Fatalf(`Wanted %s, got %s`, tt.want, jpgsToDelete)
		}
	}
}
