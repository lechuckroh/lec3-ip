package main

import (
	"io/ioutil"
	"os"
	"time"
	"path/filepath"
	"strings"
	"sort"
)

// Sort FileInfo by Name
type Files []os.FileInfo

func (files Files) Len() int {
	return len(files)
}

func (files Files) Less(i, j int) bool {
	return files[i].Name() < files[j].Name()
}

func (files Files) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}

// List image files that modified after timeAfterOptional
func ListImages(dir string, timeAfterOptional ...time.Time) ([]os.FileInfo, time.Time, error) {
	var result Files

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
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !file.ModTime().After(timeAfter) {
			continue
		}

		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			result = append(result, file)
		}
	}

	sort.Sort(result)

	return result, lastCheckTime, nil
}
