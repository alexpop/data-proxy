package log

import (
	"fmt"
)

var Level = "info"

func Debug(str string) {
	fmt.Printf(" * Logging level %s message: %s", Level, str)
}
