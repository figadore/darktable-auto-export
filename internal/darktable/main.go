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
	err := runCmd(args)
	if err != nil {
		return err
	}
	fmt.Println("Completed export to tmp file?", tmpPath)
	args = []string{"touch", "-r", params.OutputPath, tmpPath}
	runCmd(args)
	//args = []string{"cp", "-p", tmpPath, params.OutputPath}
	args = []string{"mv", tmpPath, params.OutputPath}
	runCmd(args)
	//// Update the modified time to match the existing jpg if it exists. This is what allows an album to stay intact
	//t, err := GetModifiedDate(params.OutputPath)
	//fmt.Println("Modified date:", t)
	//if err != nil {
	//	return err
	//}
	//err = os.Chtimes(tmpPath, t, t)

	//// Move tmp file to target OutputPath. This is what allows an album to stay intact
	//err = os.Rename(tmpPath, params.OutputPath)
	//if err != nil {
	//	return err
	//}
	////err = os.Chtimes(params.OutputPath, t, t)
	////if err != nil {
	////	return err
	////}
	return nil
}

func runCmd(args []string) error {
	remaining := args[1:]
	fmt.Println(args)
	cmd := exec.Command(args[0], remaining...)
	stdout, err := cmd.CombinedOutput()
	fmt.Print("=== Begin stdout/stderr ===\n", string(stdout), "\n=== End stdout/stderr ===\n")
	if err != nil {
		fmt.Println("error", err.Error())
		fmt.Println("err", err)
		return err
	}
	fmt.Println("No err from cmd")
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
