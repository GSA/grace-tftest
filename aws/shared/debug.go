package shared

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
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

// Debugf writes the formatted value of args into the
// debugWriter (default: os.Stdout)
func Debugf(format string, args ...interface{}) {
	if !isDebug {
		return
	}
	debug(fmt.Sprintf(format, args...))
}

// Debugf writes the value of data to the debugWriter
// (default: os.Stdout)
func Debug(data string) {
	if !isDebug {
		return
	}
	debug(data)
}

// Dump writes the pretty-printed values of v to the
// debugWriter (default: os.Stdout)
func Dump(v ...interface{}) {
	if !isDebug {
		return
	}
	dump(v...)
}

func debug(data string) {
	muWriter.RLock()
	_, err := spew.Fprintf(debugWriter, "%s %s\n%s", time.Now().Format(time.RFC3339), location(), data)
	muWriter.RUnlock()
	if err != nil {
		fmt.Printf("failed to write message to debug writer: %v\n", err)
		return
	}
}

func dump(v ...interface{}) {
	muWriter.RLock()
	spew.Fdump(debugWriter, v...)
	muWriter.RUnlock()
}

func location() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(4, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function)
}
