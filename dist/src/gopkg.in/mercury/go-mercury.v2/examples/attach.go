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
	clear   bool
	x86     bool
	regular bool
)

func init() {
	flag.StringVar(&mercurypath, "mercurypath", mercury.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&name, "name", "rubik", "Name of the original container")
	flag.BoolVar(&clear, "clear", false, "Attach with clear environment")
	flag.BoolVar(&x86, "x86", false, "Attach using x86 personality")
	flag.BoolVar(&regular, "regular", false, "Attach using a regular user")
	flag.Parse()
}

func main() {
	c, err := mercury.NewContainer(name, mercurypath)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	options := mercury.DefaultAttachOptions
	options.ClearEnv = false
	if clear {
		options.ClearEnv = true
	}
	if x86 {
		options.Arch = mercury.X86
	}
	if regular {
		options.UID = 1000
		options.GID = 1000
	}
	log.Printf("AttachShell\n")
	err = c.AttachShell(options)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("RunCommand\n")
	_, err = c.RunCommand([]string{"id"}, options)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
}
