package main

import (
	"flag"
	"github.com/egorperesada/chainrunner"
	"log"
)

var chainFile = flag.String("file", "", "path to file that contains chain source")

func main() {
	flag.Parse()
	if *chainFile == "" {
		log.Fatal("missed chainFile")
	}
	chain := chainrunner.FromYaml(*chainFile, false)
	chainrunner.Run(chain)
}
