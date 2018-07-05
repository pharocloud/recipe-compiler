package main

import (
	"compress/gzip"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	var input string
	var output string

	flag.StringVar(&input, "input", "", "yaml recipe file")
	flag.StringVar(&output, "output", "", "tar.gz file to output to Docker context")
	flag.Parse()

	if input == "" {
		log.Fatalln("input file is empty")
	}

	if output == "" {
		log.Fatalln("output file is empty")
	}

	i, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatalln(err)
	}
	r, err := Parse(i)
	if err != nil {
		log.Fatalln(err)
	}

	o, err := os.Create(output)
	if err != nil {
		log.Fatalln(err)
	}
	defer o.Close()
	// set up the gzip writer
	gw := gzip.NewWriter(o)
	defer gw.Close()

	err = r.WriteTo(gw)
	if err != nil {
		log.Fatalln(err)
	}
}
