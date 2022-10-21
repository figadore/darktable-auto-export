package main

import (
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
