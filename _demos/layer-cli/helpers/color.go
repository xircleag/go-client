package helpers

import (
	"fmt"
)

func ColorStringf(color int, format string, args ...interface{}) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", color, fmt.Sprintf(format, args...))
}
