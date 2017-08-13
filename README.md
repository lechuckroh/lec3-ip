Moved to https://github.com/lechuckroh/lec3

# LEC3 ImageProcessor

[![Build Status](https://travis-ci.org/lechuckroh/lec3-ip.svg?branch=master)](https://travis-ci.org/lechuckroh/lec3-ip)

Image Process module of LEC3

## Requirements

* [Go-Lang](https://golang.org/)

## Build
### Windows
* Running `build.bat` will get `bin\lec3-ip.exe`

### Linux / MacOSX
* Running `./build` will get `bin/lec3-ip`

## Test
### Windows
* Change directory to `src\lec3-ip`
* Run `go test`

### Linux / MacOSX
* Change directory to `src/lec3-ip`
* Run `go test`

## Usage
```lec3-ip [options]```

### Options
#### `-src`
Source directory of images to process.

#### `-dest`
Destination directory where processed images are stored.

#### `-watch`
Watches source directory and process new/modified images.

#### `-cfg`
Load configuration yaml file.

### Examples

```bash
lec3-ip -src=./input -dest=./output -watch=true
lec3-ip -cfg=./config/batch.yaml
```
