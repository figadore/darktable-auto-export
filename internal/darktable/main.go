package darktable

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ExportParams struct {
	Command    string // Darktable binary
	RawPath    string // Full path to raw file
	XmpPath    string // Full path to xmp (optional)
	OutputPath string // Full path to target jpg
	OnlyNew    bool   // Only export if target doesn't exist, no replace
	DryRun     bool   // Show actions that would be performed, but don't do them
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

	args := string[]{params.Command}
	args = append(args, params.RawPath)
	if params.XmpPath != "" {
		args = append(args, params.XmpPath)
	}
	tmpPath := fmt.Sprintf("%s.tmp.jpg", params.OutputPath)
	args = append(args, tmpPath)
	err := runCmd(args, params.DryRun, true)
	if err != nil {
		return err
	}
	// FIXME not sure why this won't work with os.Chtimes and os.Rename, but
	// Synology albums lost track of replaced images whenever I used a method
	// other than these commands
	args = []string{"touch", "-r", params.OutputPath, tmpPath} //FIXME check for existence first
	runCmd(args, params.DryRun, false)
	args = []string{"cp", "-p", tmpPath, params.OutputPath}
	runCmd(args, params.DryRun, false)
	args = []string{"rm", tmpPath}
	runCmd(args, params.DryRun, false)
	return nil
}

func runCmd(args []string, dryRun bool, prints bool) error {
	remaining := args[1:]
	if prints {
		fmt.Println(args)
	}
	var cmd *exec.Cmd
	if dryRun {
		cmd = exec.Command("echo", args...)
	} else {
		cmd = exec.Command(args[0], remaining...)
	}
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
