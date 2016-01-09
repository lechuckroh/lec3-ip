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

func isImage(ext string) bool {
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// List image files that modified after timeAfterOptional
func ListImages(dir string, watchDelay int, lastCheckTime time.Time) ([]os.FileInfo, time.Time, error) {
	duration := -time.Duration(watchDelay) * time.Second
	listAfter := lastCheckTime
	if watchDelay > 0 && listAfter.After(time.Unix(0, 0)) {
		listAfter = listAfter.Add(duration)
	}
	listBefore := time.Now().Add(duration)

	var result Files
	files, err := ioutil.ReadDir(dir)

	// Failed to read directory
	if err != nil {
		return result, lastCheckTime, err
	}

	lastCheckTime = time.Now()

	// Get file list that modified after EMT
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		modTime := file.ModTime()
		if !modTime.Before(listAfter) && modTime.Before(listBefore) && isImage(ext) {
			result = append(result, file)
		}
	}

	sort.Sort(result)

	if result.Len() > 0 {
		log.Printf("[ADD] %v files\n", result.Len())
	}

	return result, lastCheckTime, nil
}
