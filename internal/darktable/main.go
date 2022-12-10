package darktable

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ExportParams struct {
	Command    string
	RawPath    string
	XmpPath    string
	OutputPath string
	OnlyNew    bool
}

func Export(params ExportParams) error {
	if params.OnlyNew {
		if _, e := os.Stat(params.OutputPath); e == nil {
			fmt.Printf("jpg found at %s, skipping export\n", params.OutputPath)
			return nil
		} else if errors.Is(e, os.ErrNotExist) {
			//continue as usual
		} else {
			return e
		}
	}
	//err := sidecars.DeleteJpgIfExists(params.OutputPath)
	//if err != nil {
	//	log.Fatalf("Error deleting jpg: %v", err)
	//}
	//cmd := exec.Command("echo", params.rawPath, ":", params.xmpPath, "->", params.OutputPath)
	args := strings.Fields(params.Command)
	args = append(args, params.RawPath)
	if params.XmpPath != "" {
		args = append(args, params.XmpPath)
	}
	//args = append(args, params.OutputPath)
	tmpPath := fmt.Sprintf("%s.tmp.jpg", params.OutputPath)
	args = append(args, tmpPath)
	remaining := args[1:]
	//cmd := exec.Command("echo", remaining...)
	fmt.Println(args)
	fmt.Println(len(args))
	cmd := exec.Command(args[0], remaining...)
	//cmd := exec.Command("echo", args...)
	stdout, err := cmd.CombinedOutput()
	fmt.Print("=== Begin stdout/stderr ===\n", string(stdout), "\n=== End stdout/stderr ===")
	fmt.Println()
	if err != nil {
		fmt.Println("error", err.Error())
		fmt.Println("err", err)
		return err
	}
	// Update the modified time to match the existing jpg if it exists. This is what allows an album to stay intact
	t, err := GetModifiedDate(params.OutputPath, tmpPath)
	if err != nil {
		return err
	}

	// Move tmp file to target OutputPath. This is what allows an album to stay intact
	err = os.Rename(tmpPath, params.OutputPath)
	if err != nil {
		return err
	}
	err = os.Chtimes(params.OutputPath, t, t)
	if err != nil {
		return err
	}
	return nil
}

func GetModifiedDate(src, dst string) (time.Time, error) {
	if info, err := os.Stat(src); err == nil {
		fmt.Printf("Found existing file at target path. copying modified date %s\n", src)
		t := info.ModTime()
		return t, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// file doesn't exist, nothing to delete
		return time.Now(), nil
	} else {
		return time.Now(), err
	}
}
