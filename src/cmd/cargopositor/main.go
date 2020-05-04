package main

import (
	"compositor"
	"flag"
	"fmt"
	"os"
	"time"
)

type Flags struct {
	OutputDirectory string
	VoxelDirectory  string
	OutputTime      bool
}

var flags Flags

func init() {
	// Long format
	flag.StringVar(&flags.OutputDirectory, "output_dir", "", "output directory (default to the current path)")
	flag.StringVar(&flags.VoxelDirectory, "voxel_dir", "", "root directory for input voxel objects (default to the current path)")
	flag.BoolVar(&flags.OutputTime, "time", false, "output basic profiling information")

	// Short format
	flag.StringVar(&flags.OutputDirectory, "o", "", "shorthand for -output_dir")
	flag.StringVar(&flags.VoxelDirectory, "v", "", "shorthand for -voxel_dir")
	flag.BoolVar(&flags.OutputTime, "t", false, "shorthand for -time")
}

func main() {
	flag.Parse()

	start := time.Now()

	if flags.OutputDirectory != "" {
		if _, err := os.Stat(flags.OutputDirectory); os.IsNotExist(err) {
			if err := os.Mkdir(flags.OutputDirectory, 0755); err != nil {
				panic(err)
			}
		}
	}

	for _, batchFile := range flag.Args() {
		batch, err := compositor.FromFile(batchFile)
		if err != nil {
			fmt.Printf("could not load batch %s: %v", batchFile, err)
		} else {
			if err := batch.Run(flags.OutputDirectory, flags.VoxelDirectory); err != nil {
				fmt.Printf("could not execute batch %s: %v", batchFile, err)
			}
		}
	}

	if flags.OutputTime {
		fmt.Printf("Total time: %dms\n", time.Since(start).Milliseconds())
	}
}
