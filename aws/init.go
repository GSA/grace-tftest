package aws

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var isDebug bool

var debugWriter io.Writer
var muWriter sync.RWMutex

func init() {
	setDebugWriter(os.Stdout)

	value := os.Getenv("TFTEST_DEBUG")
	if strings.EqualFold(value, "true") {
		isDebug = true
		debug("Debugging is enabled\n")
	}
}

func setDebugWriter(w io.Writer) {
	muWriter.Lock()
	debugWriter = w
	muWriter.Unlock()
}

func debug(format string, args ...interface{}) {
	if !isDebug {
		return
	}
	muWriter.RLock()
	_, err := fmt.Fprintf(debugWriter, format, args...)
	muWriter.RUnlock()
	if err != nil {
		fmt.Printf("failed to write message to debug writer: %v\n", err)
		return
	}
}
