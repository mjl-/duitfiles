// Duitfiles is a demo program that calls duitfiles.Select and prints the selected filename.
package main

import (
	"log"

	"github.com/mjl-/duitfiles"
)

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}

func main() {
	log.SetFlags(0)

	filename, err := duitfiles.Select()
	check(err, "new files")
	log.Printf("%q\n", filename)
}
