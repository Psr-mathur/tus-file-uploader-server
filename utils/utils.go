package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetFileExtension(fileName string) string {
	return filepath.Ext(fileName)
}

func MoveFile(fileName, sourceDir, targetDir string) error {
	// Skip files with a .info extension.
	if strings.HasSuffix(fileName, ".info") {
		return fmt.Errorf("skipping file with .info extension: %s", fileName)
	}

	// Define source and target paths.
	sourcePath := filepath.Join(sourceDir, fileName)
	targetPath := filepath.Join(targetDir, fileName)

	// Ensure the target directory exists.
	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create target directory: %v", err)
	}

	// Read the source file content.
	input, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("could not read source file: %v", err)
	}

	// Write the content to the target file.
	err = os.WriteFile(targetPath, input, 0644)
	if err != nil {
		return fmt.Errorf("could not write to target file: %v", err)
	}

	// Remove the source file.
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("could not delete source file: %v", err)
	}

	log.Printf("File %s moved successfully from %s to %s\n", fileName, sourcePath, targetPath)
	return nil
}

func RemoveFileExtension(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
