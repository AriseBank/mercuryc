// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/mercury/go-mercury.v2"
)

var (
	mercurypath string
	name    string
	clear   bool
	x86     bool
	regular bool
	wg      sync.WaitGroup
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

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = io.Copy(os.Stdout, stdoutReader)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err.Error())
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = io.Copy(os.Stderr, stderrReader)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err.Error())
		}
	}()

	options := mercury.DefaultAttachOptions

	options.StdinFd = os.Stdin.Fd()
	options.StdoutFd = stdoutWriter.Fd()
	options.StderrFd = stderrWriter.Fd()

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
	if err := c.AttachShell(options); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("RunCommand\n")
	_, err = c.RunCommand([]string{"uname", "-a"}, options)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	if err = stdoutWriter.Close(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	if err = stderrWriter.Close(); err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	wg.Wait()
}
