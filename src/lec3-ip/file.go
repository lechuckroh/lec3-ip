package main

import (
	"io/ioutil"
	"os"
	"time"
)

// List image files that modified after timeAfterOptional
func ListImages(dir string, timeAfterOptional ...time.Time) ([]os.FileInfo, time.Time, error) {
	var result []os.FileInfo

	lastCheckTime := time.Now()
	files, err := ioutil.ReadDir(dir)

	// Failed to read directory
	if err != nil {
		return result, lastCheckTime, err
	}

	// Get EMT(Earliest Modified Time)
	timeAfter := time.Unix(0, 0)
	if len(timeAfterOptional) > 0 {
		timeAfter = timeAfterOptional[0]
	}

	// Get file list that modified after EMT
	for _, file := range files {
		if file.ModTime().After(timeAfter) {
			result = append(result, file)
		}
	}
	return result, lastCheckTime, nil
}
