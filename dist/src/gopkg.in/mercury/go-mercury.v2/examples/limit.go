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
	mercurypath string
	name    string
)

func init() {
	flag.StringVar(&mercurypath, "mercurypath", mercury.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&name, "name", "rubik", "Name of the container")
	flag.Parse()
}

func main() {
	c, err := mercury.NewContainer(name, mercurypath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	memLimit, err := c.MemoryLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	memorySwapLimit, err := c.MemorySwapLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	if err := c.SetMemoryLimit(memLimit / 4); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	if err := c.SetMemorySwapLimit(memorySwapLimit / 4); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
