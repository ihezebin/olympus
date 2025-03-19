package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const maximumCallerDepth = 25
const knownLoggerFrames = 3

var minimumCallerDepth int

var loggerPackage string
var loggerFileDir string

var callerInitOnce sync.Once

func getCaller() string {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			file, _ := runtime.FuncForPC(pcs[i]).FileLine(pcs[i])
			if strings.Contains(funcName, "getCaller") {
				loggerPackage = getPackageName(funcName)
				loggerFileDir = getFileDir(file)
				break
			}
		}

		minimumCallerDepth = knownLoggerFrames
	})

	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	landLogger := false
	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)
		fileDir := getFileDir(f.File)
		fileBase := filepath.Base(f.File)
		if fileDir == loggerFileDir && fileBase == "logger_test.go" {
			return fmt.Sprintf("%s:%s:%d", f.Function, fileBase, f.Line)
		}

		if pkg == loggerPackage {
			landLogger = true
			continue
		}

		// If the caller isn't part of this package, we're done
		if landLogger && pkg != loggerPackage {
			return fmt.Sprintf("%s:%s:%d", f.Function, fileBase, f.Line)
		}
	}

	// if we got here, we failed to find the caller's context
	return ""
}

func getFileDir(f string) string {
	lastSlash := strings.LastIndex(f, "/")
	if lastSlash == -1 {
		return ""
	}

	return f[:lastSlash]
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
