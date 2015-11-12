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
	"reflect"
)

//-----------------------------------------------------------------------------
// Work
//-----------------------------------------------------------------------------

type Work struct {
	dir      string
	filename string
	quit     bool
}

type Worker struct {
	workChan <-chan Work
}

func collectImages(workChan chan <- Work, finChan chan <- bool, srcDir string, watch bool) {
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
			result := filter.Run(NewFilterSource(src, work.filename))
			result.Log()

			resultImg := result.Image()
			if resultImg == nil {
				fmt.Printf("Filter result is nil. filter: %v\n", reflect.TypeOf(filter))
				break
			}

			dest = resultImg
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
	cfgFilename := flag.String("cfg", "", "configuration filename")
	srcDir := flag.String("src", "./", "source directory")
	destDir := flag.String("dest", "./output", "dest directory")
	watch := flag.Bool("watch", false, "watch directory files update")
	flag.Parse()

	// Print usage
	if flag.NFlag() == 1 && flag.Arg(1) == "help" {
		flag.Usage()
		return
	}

	// create Config
	config := NewConfig(*cfgFilename, *srcDir, *destDir, *watch)
	config.Print()

	// set maxProcess
	runtime.GOMAXPROCS(config.maxProcess)

	// Create channels
	workChan := make(chan Work, 100)
	finChan := make(chan bool)

	// WaitGroup
	wg := sync.WaitGroup{}

	// start collector
	go collectImages(workChan, finChan, config.src.dir, config.watch)

	// create filters
	deskewOption := DeskewOption{
		maxRotation: 2.0,
		incrStep: 0.2,
		threshold: 100,
		emptyLineMinDotCount: 10,
		debugOutputDir: "./debug",
		debugMode: false,
	}
	autoCropOption := AutoCropOption{
		threshold: 100,
		minRatio: 1,
		maxRatio: 3,
		maxWidthCropRate: 0.2,
		maxHeightCropRate: 0.2,
		marginTop: 5,
		marginBottom: 5,
		marginLeft: 5,
		marginRight: 5,
	}
	filters := []Filter{
		NewDeskewFilter(deskewOption),
		NewAutoCropFilter(autoCropOption),
	}

	// start workers
	for i := 0; i < config.maxProcess; i++ {
		worker := Worker{workChan}
		wg.Add(1)
		go work(worker, filters, config.dest.dir, &wg)
	}

	// wait for collector finish
	<-finChan

	// finish workers
	for i := 0; i < config.maxProcess; i++ {
		workChan <- Work{"", "", true}
	}

	wg.Wait()
}
