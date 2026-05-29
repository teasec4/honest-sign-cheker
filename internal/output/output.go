package output

import (
	"fmt"
	"os"
)

var outputFile *os.File

func Init(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	outputFile = file
	return nil
}

func Write(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Print(text)
	if outputFile != nil {
		fmt.Fprint(outputFile, text)
	}
}

func WriteLine(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Println(text)
	if outputFile != nil {
		fmt.Fprintln(outputFile, text)
	}
}

func Close() {
	if outputFile != nil {
		outputFile.Close()
	}
}
