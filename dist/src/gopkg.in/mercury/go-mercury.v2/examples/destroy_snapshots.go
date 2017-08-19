// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package main

import (
	"log"

	"gopkg.in/mercury/go-mercury.v2"
)

func main() {
	c := mercury.Containers()
	for i := range c {
		log.Printf("%s\n", c[i].Name())
		l, err := c[i].Snapshots()
		if err != nil {
			log.Fatalf("ERROR: %s\n", err.Error())
		}

		for _, s := range l {
			log.Printf("Destroying Snaphot: %s\n", s.Name)
			if err := c[i].DestroySnapshot(s); err != nil {
				log.Fatalf("ERROR: %s\n", err.Error())
			}
		}
	}
}
