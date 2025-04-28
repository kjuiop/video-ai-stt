package utils

import "path/filepath"

func GetOutputPath(outputDir, filename, targetExt string) string {
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
	newFilename := nameWithoutExt + targetExt
	return filepath.Join(outputDir, filepath.Base(newFilename))
}
