package main

import (
	"fmt"
	"flag"
	"time"
	"os"
	"image"
	"github.com/disintegration/gift"
	"path"
	"runtime"
)

func pushImages(waitChannel chan int, channel chan string, srcDir string, watch bool) {
	lastCheckTime := time.Unix(0, 0)
	var files []os.FileInfo
	var err error

	defer func() {
		channel <- ""
		waitChannel <- 1
	}()

	for {
		// List modified image files
		files, lastCheckTime, err = ListImages(srcDir, lastCheckTime)
		if err != nil {
			fmt.Println(err)
			break
		}

		// push modified image files to channel
		for _, file := range files {
			channel <- file.Name()
		}

		if watch {
			// sleep for a while
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			break
		}
	}
}

// Process image files
func processFiles(waitChannel chan int, channel chan string, srcDir string, destDir string) {
	g := gift.New(
		gift.ResizeToFit(800, 800, gift.LanczosResampling),
		gift.UnsharpMask(1.0, 1.0, 0.0),
	)

	defer func() {
		waitChannel <- 1
	}()

	for {
		filename := <-channel
		if filename == "" {
			close(channel)
			break
		}
		fmt.Printf("Processing : %+v\n", filename)

		src, err := LoadImage(path.Join(srcDir, filename))
		if err != nil {
			fmt.Printf("Error : %+v : %+v\n", filename, err)
			continue
		}

		dest := image.NewRGBA(g.Bounds(src.Bounds()))
		g.Draw(dest, src)

		// save dest Image
		err = SaveJpeg(dest, destDir, filename, 80)
		if err != nil {
			fmt.Printf("Error : %+v : %+v\n", filename, err)
			continue
		}
	}
}


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	fmt.Printf("srcDir : %+v\n", *srcDir)
	fmt.Printf("destDir : %+v\n", *destDir)
	fmt.Printf("watch : %+v\n", *watch)

	// Create channel with buffer size
	channel := make(chan string, 1000)
	waitChannel := make(chan int)

	go pushImages(waitChannel, channel, *srcDir, *watch)
	go processFiles(waitChannel, channel, *srcDir, *destDir)

	<-waitChannel
	<-waitChannel
}
