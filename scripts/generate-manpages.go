package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/app"
)

func main() {
	var outputDir string

	flag.StringVar(&outputDir, "dir", "docs/man", "directory for generated man pages")
	flag.Parse()

	if err := app.GenerateManPages(outputDir); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
