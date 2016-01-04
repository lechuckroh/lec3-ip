package main

import (
	"io/ioutil"
	"os"
	"time"
	"path/filepath"
	"strings"
	"sort"
	"log"
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
func ListImages(dir string, watchDelay int, lastCheckTime time.Time) ([]os.FileInfo, time.Time, error) {
	listAfter := lastCheckTime
	if watchDelay > 0 && lastCheckTime.After(time.Unix(0,0)) {
		listAfter = lastCheckTime.Add(-time.Duration(watchDelay) * time.Second)
	}

	var result Files
	files, err := ioutil.ReadDir(dir)

	// Failed to read directory
	if err != nil {
		return result, lastCheckTime, err
	} else {
		lastCheckTime = time.Now()
	}

	// Get file list that modified after EMT
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !file.ModTime().After(listAfter) {
			continue
		}

		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			result = append(result, file)
		}
	}

	sort.Sort(result)

	if result.Len() > 0 {
		log.Printf("%v files after %v\n", result.Len(), listAfter)
	}

	return result, lastCheckTime, nil
}
