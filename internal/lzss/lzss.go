// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package lzss

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "lzss.h"
import "C"
import (
	// "fmt"
	"sync"
	"unsafe"
)

func Encode(source, destination string) {

	var mutex sync.Mutex

	src := C.CString(source)
	defer C.free(unsafe.Pointer(src))

	dst := C.CString(destination)
	defer C.free(unsafe.Pointer(dst))

	mutex.Lock()
	C.encode_file(src, dst)
	mutex.Unlock()
}
