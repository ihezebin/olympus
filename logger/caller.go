package logger

import (
	"fmt"
	"runtime"
)

func getCaller(skipFrameCount int) string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skipFrameCount+1, pc)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}
