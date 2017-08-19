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
)

func init() {
	flag.StringVar(&mercurypath, "mercurypath", mercury.DefaultConfigPath(), "Use specified container path")
	flag.Parse()
}

func main() {
	log.Printf("Defined containers:\n")
	c := mercury.DefinedContainers(mercurypath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}

	log.Println()

	log.Printf("Active containers:\n")
	c = mercury.ActiveContainers(mercurypath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}

	log.Println()

	log.Printf("Active and Defined containers:\n")
	c = mercury.ActiveContainers(mercurypath)
	for i := range c {
		log.Printf("%s (%s)\n", c[i].Name(), c[i].State())
	}
}
