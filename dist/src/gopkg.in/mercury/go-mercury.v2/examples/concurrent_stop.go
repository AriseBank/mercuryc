// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"flag"
	"log"
	"runtime"
	"strconv"
	"sync"

	"gopkg.in/mercury/go-mercury.v2"
)

var (
	mercurypath string
	count   int
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&mercurypath, "mercurypath", mercury.DefaultConfigPath(), "Use specified container path")
	flag.IntVar(&count, "count", 10, "Number of containers")
	flag.Parse()
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			c, err := mercury.NewContainer(strconv.Itoa(i), mercurypath)
			if err != nil {
				log.Fatalf("ERROR: %s\n", err.Error())
			}

			log.Printf("Stoping the container (%d)...\n", i)
			if err := c.Stop(); err != nil {
				log.Fatalf("ERROR: %s\n", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
