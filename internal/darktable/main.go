package darktable

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type ExportParams struct {
	Command    string // Darktable binary
	RawPath    string // Full path to raw file
	XmpPath    string // Full path to xmp (optional)
	OutputPath string // Full path to target jpg
	OnlyNew    bool   // Only export if target doesn't exist, no replace
	DryRun     bool   // Show actions that would be performed, but don't do them
}

func copyToLocal(src, dst string) error {
	fmt.Printf("Copy from '%s' to '%s'\n", src, dst)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, data, 0666)
	if err != nil {
		return err
	}
	mTime, err := GetModifiedDate(src)
	if err != nil {
		return err
	}
	fmt.Printf("Updating modified time of '%s'\n", dst)
	os.Chtimes(dst, mTime, mTime)
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(path string) error {
	err := os.Remove(path)
	return err
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

	args := []string{params.Command}
	h := sha256.New()
	h.Write([]byte(params.RawPath))
	bs := h.Sum(nil)

	localRawPath := fmt.Sprintf("/tmp/%x.arw", bs[:4])
	err := copyToLocal(params.RawPath, localRawPath)
	if err != nil {
		return err
	}
	defer deleteFile(localRawPath)
	args = append(args, localRawPath)
	if params.XmpPath != "" {
		args = append(args, params.XmpPath)
	}
	mTime, err := GetModifiedDate(params.OutputPath)
	if err != nil {
		return err
	}
	// Write to tmp file since darktable cli creates a new file if target
	// filename exists
	tmpPath := fmt.Sprintf("%s.tmp.jpg", filepath.ToSlash(params.OutputPath))
	// escape slashes on windows
	args = append(args, tmpPath)
	err = runCmd(args, params.DryRun, true)
	if err != nil {
		return err
	}
	// Try to edit in place, preserving existing photos so Synology doesn't
	// remove them from albums
	// FIXME not sure why this won't work with os.Chtimes and os.Rename, but
	// Synology albums lost track of replaced images whenever I used a method
	// other than these commands
	os.Rename(tmpPath, params.OutputPath)
	os.Chtimes(params.OutputPath, mTime, mTime)

	//args = []string{"touch", "-r", params.OutputPath, tmpPath} //FIXME check for existence first
	//runCmd(args, params.DryRun, false)
	//args = []string{"cp", "-p", tmpPath, params.OutputPath}
	//runCmd(args, params.DryRun, false)
	//args = []string{"rm", tmpPath}
	//runCmd(args, params.DryRun, false)
	return nil
}

func runCmd(args []string, dryRun bool, prints bool) error {
	remaining := args[1:]
	if prints {
		fmt.Println(args)
	}
	var cmd *exec.Cmd
	if dryRun {
		fmt.Println(args)
		return nil
	} else {
		cmd = exec.Command(args[0], remaining...)
		stdout, err := cmd.CombinedOutput()
		if len(stdout) != 0 {
			if !dryRun {
				fmt.Print("=== Begin stdout/stderr ===\n", string(stdout), "\n=== End stdout/stderr ===\n")
			} else if prints {
				fmt.Print(string(stdout))
			}
		}
		if err != nil {
			fmt.Println("cmd error", err.Error())
			fmt.Println("cmd err", err)
			return err
		}
		return nil
	}
}

func GetModifiedDate(src string) (time.Time, error) {
	if info, err := os.Stat(src); err == nil {
		fmt.Printf("Found existing file at target path. copying modified date %s\n", src)
		t := info.ModTime()
		return t, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// file doesn't exist, nothing to delete
		fmt.Printf("Did not find existing file at target path. Can't copy modified date %s\n", src)
		return time.Now(), nil
	} else {
		return time.Now(), err
	}
}
