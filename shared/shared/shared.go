// Package shared contains everything shared between all other packages.
package shared

import (
	"log"
	"os"
)

const logFlags = log.LstdFlags | log.Lshortfile

var (
	// InfoLogger logs on INFO level.
	InfoLogger = log.New(os.Stdout, "INFO ", logFlags)
	// ErrorLogger logs on ERROR level.
	ErrorLogger = log.New(os.Stderr, "ERROR ", logFlags)
)
