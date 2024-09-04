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

	fmt.Println("Create temporary directory")
	dirpath, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create a temporary directory.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	deviantFetch := deviantart.FetchFavorites(username, dirpath, 4)
	shared.InfoLogger.Println("Deviations fetched.")
	err = deviantart.SaveJSON(deviantFetch, "deviantFetch.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}
