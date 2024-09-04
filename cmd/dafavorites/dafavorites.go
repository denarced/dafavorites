// Package main.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/denarced/dafavorites/shared/deviantart"
	"github.com/denarced/dafavorites/shared/shared"
)

func main() {
	shared.InitLogging()
	shared.Logger.Info("Start.", "args", os.Args)
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Missing username")
		fmt.Printf("Usage: %s {username}\n", os.Args[0])
		os.Exit(4)
		return
	}

	username := strings.TrimSpace(os.Args[1])
	if len(username) == 0 {
		fmt.Println("Username can't be empty")
		os.Exit(1)
	}

	shared.Logger.Debug("Create temporary directory.")
	dirpath, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create a temporary directory.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	deviantFetch := deviantart.FetchFavorites(username, dirpath, 4)
	shared.Logger.Info("Deviations fetched.", "count", len(deviantFetch.SavedDeviations))
	err = deviantart.SaveJSON(deviantFetch, "deviantFetch.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed.")
		fmt.Fprintln(os.Stderr, err)
		shared.Logger.Error("Done, failed.", "error", err)
		os.Exit(3)
	}
	fmt.Printf("Done. Deviations download to %s.\n", dirpath)
	shared.Logger.Info("Done.")
}
