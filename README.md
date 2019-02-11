# Web Crawler
A simple web crawler built using Go (powered by go routines)

## Demo
[![asciicast](https://asciinema.org/a/VUzIgdqhWtAFqkwrWoOHeVxja.svg)](https://asciinema.org/a/VUzIgdqhWtAFqkwrWoOHeVxja)

## How to build the binary
1. Run `go get -v` to download the dependencies
2. Run `go build -o webcrawler` to build the binary

## Usage
`./webcrawler -baseurl https://golang.org -max-depth 2`

This will start the webcrawler and generate two files
1. `url-tree.txt` which shows the links between pages. The default file name can
   be changed by `-tree-file-name` flag. (You can disable the tree generation by
   setting `-show-tree` flag to `false`).
2. `sitemap.xml` which contains the sitemap in xml format.

## How to run tests
```go
go test -v ./...
```

## Things that can be improved
1. Command line flags: The flag handling can be improved by using https://github.com/spf13/cobra
2. Configuration: It would be nice to have some configuration management (It
   could be done by https://github.com/spf13/viper)
3. Performance: All the go routines write to a single shared state. The
   performance might improve if we use channels. (we will have to benchmark it
   to find the actual performance improvements)

## Issue with Page Links Tree
Please note that the tree generated shows children for any given node *only
once*
For example, if there is a page with URLs links as
```
foo
 |-bar
      |-lorem
            |-ipsum
 |-lorem
      |-ipsum
```
the generated tree would look like
```
foo
 |-bar
      |-lorem
            |-ipsum
 |-lorem
```
That is, the child nodes of any given node are shown only once.