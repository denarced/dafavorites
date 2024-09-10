// Package main.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"

	"github.com/denarced/dafavorites/lib/dafavorites"
	"github.com/denarced/dafavorites/shared/shared"
	"github.com/spf13/afero"
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

	ctx := newProductionContext(&afero.Afero{Fs: afero.NewOsFs()}, username)
	deviantFetch := dafavorites.FetchFavorites(dirpath, 4, ctx)
	shared.Logger.Info("Deviations fetched.", "count", len(deviantFetch.SavedDeviations))
	err = dafavorites.SaveJSON(deviantFetch, filepath.Join(dirpath, "deviantFetch.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed.")
		fmt.Fprintln(os.Stderr, err)
		shared.Logger.Error("Done, failed.", "error", err)
		os.Exit(3)
	}
	fmt.Printf("Done. Deviations downloaded to %s.\n", dirpath)
	shared.Logger.Info("Done.")
}

type productionContext struct {
	fsys     *afero.Afero
	username string
}

func newProductionContext(fsys *afero.Afero, username string) *productionContext {
	return &productionContext{
		fsys:     fsys,
		username: username,
	}
}

// Username .
func (v *productionContext) Username() string {
	return v.username
}

func (v *productionContext) Fsys() *afero.Afero {
	return v.fsys
}

func (*productionContext) CreateClient() dafavorites.HTTPClient {
	return newRealHTTPClient()
}

// RealHTTPClient implements deviantart.HTTPClient.
type RealHTTPClient struct {
	client *http.Client
}

func newRealHTTPClient() *RealHTTPClient {
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	return &RealHTTPClient{client: client}
}

// Fetch .
func (v *RealHTTPClient) Fetch(url string) ([]byte, error) {
	res, err := v.client.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}
