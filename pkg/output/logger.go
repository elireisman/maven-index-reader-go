package output

import (
	"fmt"
	"log"

	"github.com/elireisman/maven-index-reader-go/pkg/config"
	"github.com/elireisman/maven-index-reader-go/pkg/data"
)

type Logger struct {
	logger *log.Logger
	input  <-chan data.Record
}

func NewLogger(l *log.Logger, in <-chan data.Record, _ config.Index) Logger {
	l.Printf("Output: printing data.Record structs to stdout...")
	return Logger{l, in}
}

func (l Logger) Write() error {
	count := 0
	for record := range l.input {
		fmt.Printf("%+v\n", record)
		count++
	}

	l.logger.Printf("Logger: completed print of %d records", count)
	return nil
}
