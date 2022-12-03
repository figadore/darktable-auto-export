package darktable

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/figadore/darktable-auto-export/internal/sidecars"
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
	err := sidecars.DeleteJpgIfExists(params.OutputPath)
	if err != nil {
		log.Fatalf("Error deleting jpg: %v", err)
	}
	//cmd := exec.Command("echo", params.rawPath, ":", params.xmpPath, "->", params.OutputPath)
	args := strings.Fields(params.Command)
	args = append(args, params.RawPath)
	if params.XmpPath != "" {
		args = append(args, params.XmpPath)
	}
	args = append(args, params.OutputPath)
	remaining := args[1:]
	//cmd := exec.Command("echo", remaining...)
	fmt.Println(args)
	fmt.Println(len(args))
	cmd := exec.Command(args[0], remaining...)
	//cmd := exec.Command("echo", args...)
	stdout, err := cmd.CombinedOutput()
	fmt.Println("stdout/stderr:", string(stdout), ":end stdout/stderr;")
	fmt.Println()
	if err != nil {
		fmt.Println("error", err.Error())
		fmt.Println("err", err)
		return err
	}
	return nil
}
