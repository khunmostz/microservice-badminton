package logger

import (
	"log"
)

func L() *log.Logger { return log.Default() }
