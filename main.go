package main

import (
	"os"

	"github.com/sleepy-day/sqline/texteditor"
)

func main() {
	f, err := os.ReadFile("testfile.txt")
	if err != nil {
		panic(err)
	}
	texteditor.Start(f)
}
