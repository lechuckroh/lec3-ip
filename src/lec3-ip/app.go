package main

import (
	"fmt"
	"flag"
	"time"
	"os"
	"image"
	"path"
	"runtime"
	"sync"
)

type Work struct {
	dir string
	filename string
	quit bool
}

type Worker struct {
	workChan <-chan Work
}

func collectImages(workChan chan<- Work, finChan chan<- bool, srcDir string, watch bool) {
	defer func() {
		finChan <- true
	}()

	lastCheckTime := time.Unix(0, 0)
	var files []os.FileInfo
	var err error

	for {
		// List modified image files
		files, lastCheckTime, err = ListImages(srcDir, lastCheckTime)
		if err != nil {
			fmt.Println(err)
			break
		}

		// add works
		for _, file := range files {
			workChan <- Work{srcDir, file.Name(), false}
		}

		if watch {
			// sleep for a while
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			break
		}
	}
}

func work(worker Worker, filters []Filter, destDir string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	for {
		work := <-worker.workChan
		if work.quit {
			break
		}

		fmt.Printf("Processing : %+v\n", work.filename)

		src, err := LoadImage(path.Join(work.dir, work.filename))
		if err != nil {
			fmt.Printf("Error : %+v : %+v\n", work.filename, err)
			continue
		}

		// run filters
		var dest image.Image
		for _, filter := range filters {
			dest = filter.run(src)
			src = dest
		}

		// save dest Image
		err = SaveJpeg(dest, destDir, work.filename, 80)
		if err != nil {
			fmt.Printf("Error : %+v : %+v\n", work.filename, err)
			continue
		}
	}
}

func main() {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

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

	// Create channels
	workChan := make(chan Work, 100)
	finChan := make(chan bool)

	// WaitGroup
	wg := sync.WaitGroup{}

	// start collector
	go collectImages(workChan, finChan, *srcDir, *watch)

	// create filters
	filters := []Filter{
		NewEdgeDetectionFilter(),
	}

	// start workers
	for i := 0; i < numCpu; i++ {
		worker := Worker{workChan}
		wg.Add(1)
		go work(worker, filters, *destDir, &wg)
	}

	// wait for collector finish
	<- finChan

	// finish workers
	for i := 0; i < numCpu; i++ {
		workChan <- Work{"", "", true}
	}

	wg.Wait()
}
