// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"
	"time"

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

	c.SetLogFile("/tmp/" + name + ".log")
	c.SetLogLevel(mercury.TRACE)

	log.Printf("Starting the container...\n")
	if err := c.Start(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("Waiting container to startup networking...\n")
	if _, err := c.WaitIPAddresses(5 * time.Second); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
