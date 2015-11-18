package main
import (
	"fmt"
	"runtime"
	"github.com/olebedev/config"
	"flag"
	"log"
)

type SrcOption struct {
	dir       string
	recursive bool
}
type DestOption struct {
	dir string
}

type FilterOption struct {
	name   string
	filter Filter
}

type Config struct {
	src        SrcOption
	dest       DestOption
	watch      bool
	maxProcess int
	filters    []FilterOption
}

func (c *Config) LoadYaml(filename string) {
	cfg, err := config.ParseYamlFile(filename)
	if err != nil {
		log.Printf("Error : Failed to parse %v : %v\n", filename, err)
		return
	}

	fmt.Printf("Loading %v\n", filename)

	c.src.dir = cfg.UString("src.dir", "")
	c.src.recursive = cfg.UBool("src.recursive", false)
	c.dest.dir = cfg.UString("dest.dir", "")
	c.watch = cfg.UBool("watch", false)
	c.maxProcess = cfg.UInt("maxProcess", 0)
	if c.maxProcess <= 0 {
		c.maxProcess = runtime.NumCPU()
	}

	// Load filters
	for i := 0;; i++ {
		m, err := cfg.Map(fmt.Sprintf("filters.%v", i))
		if err != nil {
			break
		}
		name, ok := m["name"]
		if !ok {
			continue
		}

		options := m["options"].(map[string]interface{})

		switch name {
		case "deskew":
			if option, err := NewDeskewOption(options); err == nil {
				filter := NewDeskewFilter(*option)
				filterOption := FilterOption{
					name: name.(string),
					filter: filter,
				}
				c.filters = append(c.filters, filterOption)
			} else {
				log.Printf("Failed to read filter : %v : %v\n", name, err)
			}
		case "autoCrop":
			if option, err := NewAutoCropOption(options); err == nil {
				filter := NewAutoCropFilter(*option)
				filterOption := FilterOption{
					name: name.(string),
					filter: filter,
				}
				c.filters = append(c.filters, filterOption)
			} else {
				log.Printf("Failed to read filter : %v : %v\n", name, err)
			}
		default:
			log.Printf("Unhandled filter name : %v\n", name)
		}
	}
}

func (c *Config) Print() {
	fmt.Printf("src.dir : %v\n", c.src.dir)
	fmt.Printf("dest.dir : %v\n", c.dest.dir)
	fmt.Printf("watch : %v\n", c.watch)
	fmt.Printf("maxProcess : %v\n", c.maxProcess)
	fmt.Printf("filters : %v\n", len(c.filters))
}

func NewConfig(cfgFilename string, srcDir string, destDir string, watch bool) *Config {
	config := Config{}

	if cfgFilename != "" {
		config.LoadYaml(cfgFilename)
	} else {
		// overwrite config with command line options
		if srcFlag := flag.Lookup("src"); srcFlag != nil {
			config.src.dir = srcDir
		}
		if destFlag := flag.Lookup("dest"); destFlag != nil {
			config.dest.dir = destDir
		}
		if watchFlag := flag.Lookup("watch"); watchFlag != nil {
			config.watch = watch
		}
	}

	return &config
}
