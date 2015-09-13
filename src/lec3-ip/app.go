package main

import "fmt"

func main() {
	files, _, err := ListImages("./")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		fmt.Printf("%+v\n", file.Name())
	}
}
