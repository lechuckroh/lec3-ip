package main

import (
	"fmt"
	"flag"
	"time"
	"os"
)

func pushImages(channel chan os.FileInfo, srcDir string, watch bool) {
	lastCheckTime := time.Unix(0, 0)
	var files []os.FileInfo
	var err error

	defer func() {
		channel <- nil
	}()

	for {
		// List modified image files
		files, lastCheckTime, err = ListImages(srcDir, lastCheckTime)
		if err != nil {
			fmt.Println(err)
			return
		}

		// push modified image files to channel
		for _, file := range files {
			channel <- file
		}

		if watch {
			// sleep for a while
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			return
		}
	}
}

// Process image files
func processFiles(channel chan os.FileInfo, destDir string) {
	for {
		file := <-channel
		if file == nil {
			break
		}
		fmt.Printf("%+v\n", file.Name())
	}
}


func main() {
	// Parse command-line options
	srcDir := flag.String("src", "./", "source directory")
	destDir := flag.String("dest", "./output", "dest directory")
	watch := flag.Bool("watch", false, "watch directory files update")

	flag.Parse()

	// Print usage
	if flag.NFlag() == 1 && flag.Arg(1) == "help" {
		flag.Usage()
		return
	}

	// Create channel
	channel := make(chan os.FileInfo)

	go pushImages(channel, *srcDir, *watch)

	processFiles(channel, *destDir)
}
