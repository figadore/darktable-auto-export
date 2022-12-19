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
	args := strings.Fields(params.Command)
	args = append(args, params.RawPath)
	if params.XmpPath != "" {
		args = append(args, params.XmpPath)
	}
	//args = append(args, params.OutputPath)
	tmpPath := fmt.Sprintf("%s.tmp.jpg", params.OutputPath)
	args = append(args, tmpPath)
	// Uncomment this line to do a dry run (maybe turn this into a param, but be sure to include dry run deleting files)
	//args = append([]string{"echo"}, args...)
	err := runCmd(args)
	if err != nil {
		return err
	}
	// FIXME not sure why this won't work with os.Chtimes and os.Rename, but
	// Synology albums lost track of replaced images whenever I used a method
	// other than these commands
	//fmt.Println("Completed export to tmp file?", tmpPath)
	args = []string{"touch", "-r", params.OutputPath, tmpPath} //FIXME check for existence first
	runCmd(args)
	args = []string{"cp", "-p", tmpPath, params.OutputPath}
	runCmd(args)
	args = []string{"rm", tmpPath}
	runCmd(args)
	return nil
}

func runCmd(args []string) error {
	remaining := args[1:]
	fmt.Println(args)
	cmd := exec.Command(args[0], remaining...)
	stdout, err := cmd.CombinedOutput()
	if len(stdout) != 0 {
		fmt.Print("=== Begin stdout/stderr ===\n", string(stdout), "\n=== End stdout/stderr ===\n")
	}
	if err != nil {
		fmt.Println("cmd error", err.Error())
		fmt.Println("cmd err", err)
		return err
	}
	return nil
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
