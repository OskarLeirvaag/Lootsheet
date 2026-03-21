// Package main is the entry point for the LootSheet CLI, a double-entry
// bookkeeping tool for tabletop RPG parties.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/OskarLeirvaag/Lootsheet/src/app"
)

func main() {
	if err := app.Run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
