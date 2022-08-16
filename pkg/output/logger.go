package output

import (
	"fmt"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/data"
)

type Logger struct {
	logger *log.Logger
	input  <-chan data.Record
}

func NewLogger(l *log.Logger, in <-chan data.Record) Logger {
	return Logger{l, in}
}

func (l Logger) Write() error {
	l.logger.Printf("Logger: printing data.Record structs to stdout...\n")

	count := 0
	for record := range l.input {
		fmt.Printf("%+v\n", record)
		count++
	}

	l.logger.Printf("Logger: completed print of %d records\n", count)
	return nil
}
