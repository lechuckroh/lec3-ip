# LEC3 ImageProcessor

Image Process module of LEC3

## Requirements

* [Go-Lang](https://golang.org/)

## Build

### Windows

* Running `build.bat` will get `bin\lec3-ip.exe`

### Linux / MacOSX

* Running `./build` will get `bin/lec3-ip`

## Usage

```lec3-ip [options]```

### Options

#### `-src`
Source directory of images to process.

#### `-dest`
Destination directory where processed images are stored.

#### `-watch`
Watches source directory and process new/modified images.

### Examples

```bash
lec3-ip -src=./input -dest=./output =watch=true
```