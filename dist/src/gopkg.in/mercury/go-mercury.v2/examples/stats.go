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

	// mem
	memUsed, err := c.MemoryUsage()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("MemoryUsage: %s\n", memUsed)
	}

	memLimit, err := c.MemoryLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("MemoryLimit: %s\n", memLimit)
	}

	// kmem
	kmemUsed, err := c.KernelMemoryUsage()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("KernelMemoryUsage: %s\n", kmemUsed)
	}

	kmemLimit, err := c.KernelMemoryLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("KernelMemoryLimit: %s\n", kmemLimit)
	}

	// swap
	swapUsed, err := c.MemorySwapUsage()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("MemorySwapUsage: %s\n", swapUsed)
	}

	swapLimit, err := c.MemorySwapLimit()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("MemorySwapLimit: %s\n", swapLimit)
	}

	// blkio
	blkioUsage, err := c.BlkioUsage()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	} else {
		log.Printf("BlkioUsage: %s\n", blkioUsage)
	}

	cpuTime, err := c.CPUTime()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	log.Printf("cpuacct.usage: %s\n", cpuTime)

	cpuTimePerCPU, err := c.CPUTimePerCPU()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	log.Printf("cpuacct.usageerrpercpu: %v\n", cpuTimePerCPU)

	cpuStats, err := c.CPUStats()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	log.Printf("cpuacct.stat: %v\n", cpuStats)

	interfaceStats, err := c.InterfaceStats()
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	log.Printf("InterfaceStats: %v\n", interfaceStats)
}
