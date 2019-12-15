package main

import (
	"chainrunner"
	"flag"
	"log"
)

var chainFile = flag.String("file", "", "path to file that contains chain source")

func main() {
	if *chainFile == "" {
		log.Fatal("missed chainFile")
	}
	flag.Parse()
	chain := chainrunner.FromYaml(*chainFile, false)
	chainrunner.Run(chain)
}
