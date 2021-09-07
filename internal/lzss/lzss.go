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
