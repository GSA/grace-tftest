package aws

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var is_debug bool

var debug_writer io.Writer
var muWriter sync.RWMutex

func init() {
	muWriter.Lock()
	debug_writer = os.Stdout
	muWriter.Unlock()

	value := os.Getenv("TFTEST_DEBUG")
	if strings.EqualFold(value, "true") {
		is_debug = true
	}
}

func setDebugWriter(w io.Writer) {
	muWriter.Lock()
	debug_writer = w
	muWriter.Unlock()
}

func debug(format string, args ...interface{}) {
	if !is_debug {
		return
	}
	muWriter.RLock()
	_, err := fmt.Fprintf(debug_writer, format, args...)
	muWriter.RUnlock()
	if err != nil {
		fmt.Printf("failed to write message to debug writer: %v\n", err)
		return
	}
}
