// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"

	"gopkg.in/mercury/go-mercury.v2"
)

var (
	mercurypath   string
	directory string
	name      string
	stop      bool
	verbose   bool
)

func init() {
	flag.StringVar(&mercurypath, "mercurypath", mercury.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&name, "name", "rubik", "Name of the container")
	flag.StringVar(&directory, "directory", "/tmp/rubik", "directory to save the checkpoint in")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&stop, "stop", false, "Stop the container after checkpointing.")
	flag.Parse()
}

func main() {
	c, err := mercury.NewContainer(name, mercurypath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	if verbose {
		c.SetVerbosity(mercury.Verbose)
	}

	options := mercury.CheckpointOptions{
		Directory: directory,
		Verbose:   verbose,
		Stop:      stop,
	}

	if err := c.Checkpoint(options); err != nil {
		log.Printf("ERROR: %s\n", err.Error())
	}
}
