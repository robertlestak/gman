package web

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

//go:embed web/*
var webContent embed.FS

func WriteWebContent(dir string) error {
	l := log.WithField("fn", "writeWebContent")
	l.Debug("writing web content")

	// Create the target directory
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Use a function to recursively walk the embedded filesystem and write files out
	var writeFunc fs.WalkDirFunc = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dir, path)

		if d.IsDir() {
			// Create the directory on disk
			return os.MkdirAll(targetPath, 0755)
		} else {
			// Read the embedded file
			data, err := fs.ReadFile(webContent, path)
			if err != nil {
				return err
			}

			// Write the file to disk
			return os.WriteFile(targetPath, data, 0644)
		}
	}
	if err := fs.WalkDir(webContent, "web", writeFunc); err != nil {
		return err
	}
	// move the contents of web to the root of the dir, including hidden files
	if err := moveFilesToParentDir(filepath.Join(dir, "web")); err != nil {
		return err
	}
	return nil
}

func moveFilesToParentDir(srcDir string) error {
	// Get the parent directory
	parentDir := filepath.Dir(srcDir)

	// Read the list of files and directories in the source directory
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(srcDir, file.Name())
		destPath := filepath.Join(parentDir, file.Name())

		// Move the item to the parent directory
		err := os.Rename(srcPath, destPath)
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", srcPath, destPath, err)
		}
	}

	// Optionally, remove the source directory if it's empty
	// Note: This step can be removed if you want to keep the source directory
	_ = os.Remove(srcDir)

	return nil
}
