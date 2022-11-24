/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// unlockCmd represents the unlock command
var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Clear darktable db lock files",
	Long: `Only clear the darktable db lock files if darktable
is definitely not in use.`,
	RunE: Unlock,
}

func init() {
	rootCmd.AddCommand(unlockCmd)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to read home directory")
	}
	defaultConfigDir := filepath.Join(home, ".var/app/org.darktable.Darktable/config/darktable/")
	unlockCmd.Flags().StringVarP(&unlockOpts.lockDir, "lockdir", "", defaultConfigDir, "Directory where darktable lock files are kept. Often ~/.config/darktable for local installations")
}

var unlockOpts unlockOptions

type unlockOptions struct {
	lockDir string
}

func Unlock(cmd *cobra.Command, args []string) error {
	fmt.Println("Deleting lock files")

	dbLockPath := filepath.Join(unlockOpts.lockDir, "data.db.lock")
	libraryLockPath := filepath.Join(unlockOpts.lockDir, "library.db.lock")
	err := deleteFile(dbLockPath)
	if err != nil {
		return err
	}
	err = deleteFile(libraryLockPath)
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Removing %s\n", path)
		e := os.Remove(path)
		if e != nil {
			return e
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// file doesn't exist, nothing to delete
		fmt.Println("No file found at", path)
		return nil
	} else {
		return err
	}
	return nil
}
