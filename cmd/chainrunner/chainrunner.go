package main

import (
	"flag"
	"github.com/egorperesada/chainrunner"
	"log"
)

var chainFile = flag.String("file", "", "path to file that contains chain source")

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *chainFile == "" {
		log.Fatal("missed chainFile")
	}
	provider, err := chainrunner.NewYamlProviderFromFile(*chainFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(provider.CreateChain().Execute())
}
