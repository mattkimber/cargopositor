package main

import (
	"compositor"
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	for _, batchFile := range flag.Args() {
		batch, err := compositor.FromFile(batchFile)
		if err != nil {
			fmt.Errorf("could not load batch %s: %v", batchFile, err)
		} else {
			if err := batch.Run(); err != nil {
				fmt.Errorf("could not execute batch %s: %v", batchFile, err)
			}
		}
	}
}
