// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package mercury

// #cgo pkg-config: mercury
// #cgo LDFLAGS: -lmercury -lutil
// #include <mercury/mercurycontainer.h>
// #include <mercury/version.h>
// #include "mercury-binding.h"
// #ifndef MERCURY_DEVEL
// #define MERCURY_DEVEL 0
// #endif
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// NewContainer returns a new container struct.
func NewContainer(name string, mercurypath ...string) (*Container, error) {
	var container *C.struct_mercury_container

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if mercurypath != nil && len(mercurypath) == 1 {
		cmercurypath := C.CString(mercurypath[0])
		defer C.free(unsafe.Pointer(cmercurypath))

		container = C.mercury_container_new(cname, cmercurypath)
	} else {
		container = C.mercury_container_new(cname, nil)
	}

	if container == nil {
		return nil, ErrNewFailed
	}
	c := &Container{container: container, verbosity: Quiet}

	// http://golang.org/pkg/runtime/#SetFinalizer
	runtime.SetFinalizer(c, Release)
	return c, nil
}

// Acquire increments the reference counter of the container object.
func Acquire(c *Container) bool {
	return C.mercury_container_get(c.container) == 1
}

// Release decrements the reference counter of the container object.
func Release(c *Container) bool {
	// http://golang.org/pkg/runtime/#SetFinalizer
	runtime.SetFinalizer(c, nil)

	return C.mercury_container_put(c.container) == 1
}

// Version returns the MERCURY version.
func Version() string {
	version := C.GoString(C.mercury_get_version())
	if C.MERCURY_DEVEL == 1 {
		fmt.Sprintf("%s (devel)", version)
	}
	return version
}

// GlobalConfigItem returns the value of the given global config key.
func GlobalConfigItem(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.GoString(C.mercury_get_global_config_item(cname))
}

// DefaultConfigPath returns default config path.
func DefaultConfigPath() string {
	return GlobalConfigItem("mercury.mercurypath")
}

// DefaultLvmVg returns the name of the default LVM volume group.
func DefaultLvmVg() string {
	return GlobalConfigItem("mercury.bdev.lvm.vg")
}

// DefaultZfsRoot returns the name of the default ZFS root.
func DefaultZfsRoot() string {
	return GlobalConfigItem("mercury.bdev.zfs.root")
}

// ContainerNames returns the names of defined and active containers on the system.
func ContainerNames(mercurypath ...string) []string {
	var size int
	var cnames **C.char

	if mercurypath != nil && len(mercurypath) == 1 {
		cmercurypath := C.CString(mercurypath[0])
		defer C.free(unsafe.Pointer(cmercurypath))

		size = int(C.list_all_containers(cmercurypath, &cnames, nil))
	} else {

		size = int(C.list_all_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// Containers returns the defined and active containers on the system. Only
// containers that could retrieved successfully are returned.
func Containers(mercurypath ...string) []Container {
	var containers []Container

	for _, v := range ContainerNames(mercurypath...) {
		if container, err := NewContainer(v, mercurypath...); err == nil {
			containers = append(containers, *container)
		}
	}

	return containers
}

// DefinedContainerNames returns the names of the defined containers on the system.
func DefinedContainerNames(mercurypath ...string) []string {
	var size int
	var cnames **C.char

	if mercurypath != nil && len(mercurypath) == 1 {
		cmercurypath := C.CString(mercurypath[0])
		defer C.free(unsafe.Pointer(cmercurypath))

		size = int(C.list_defined_containers(cmercurypath, &cnames, nil))
	} else {

		size = int(C.list_defined_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// DefinedContainers returns the defined containers on the system.  Only
// containers that could retrieved successfully are returned.
func DefinedContainers(mercurypath ...string) []Container {
	var containers []Container

	for _, v := range DefinedContainerNames(mercurypath...) {
		if container, err := NewContainer(v, mercurypath...); err == nil {
			containers = append(containers, *container)
		}
	}

	return containers
}

// ActiveContainerNames returns the names of the active containers on the system.
func ActiveContainerNames(mercurypath ...string) []string {
	var size int
	var cnames **C.char

	if mercurypath != nil && len(mercurypath) == 1 {
		cmercurypath := C.CString(mercurypath[0])
		defer C.free(unsafe.Pointer(cmercurypath))

		size = int(C.list_active_containers(cmercurypath, &cnames, nil))
	} else {

		size = int(C.list_active_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// ActiveContainers returns the active containers on the system. Only
// containers that could retrieved successfully are returned.
func ActiveContainers(mercurypath ...string) []Container {
	var containers []Container

	for _, v := range ActiveContainerNames(mercurypath...) {
		if container, err := NewContainer(v, mercurypath...); err == nil {
			containers = append(containers, *container)
		}
	}

	return containers
}

func VersionNumber() (major int, minor int) {
	major = C.MERCURY_VERSION_MAJOR
	minor = C.MERCURY_VERSION_MINOR

	return
}

func VersionAtLeast(major int, minor int, micro int) bool {
	if C.MERCURY_DEVEL == 1 {
		return true
	}

	if major > C.MERCURY_VERSION_MAJOR {
		return false
	}

	if major == C.MERCURY_VERSION_MAJOR &&
		minor > C.MERCURY_VERSION_MINOR {
		return false
	}

	if major == C.MERCURY_VERSION_MAJOR &&
		minor == C.MERCURY_VERSION_MINOR &&
		micro > C.MERCURY_VERSION_MICRO {
		return false
	}

	return true
}

func IsSupportedConfigItem(key string) bool {
	configItem := C.CString(key)
	defer C.free(unsafe.Pointer(configItem))
	return bool(C.go_mercury_config_item_is_supported(configItem))
}
