package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/denarced/dafavorites/deviantart"
)

const baseRss = "http://backend.deviantart.com/rss.xml?q=favby%3A___usern___%2F2573873&type=deviation"

func fetchItems(username string) (rssItems []deviantart.RssItem, err error) {
	log.Printf("Fetch information on %s's favorite deviations", username)
	// Fetch first RssFile
	// Grab all items and store them
	// If RssFile.NextUrl contains something, continue with it
	url := strings.Replace(baseRss, "___usern___", username, 1)
	var rssFile deviantart.RssFile
	rssFile, err = deviantart.FetchRssFile(url)
	if err != nil {
		return
	}
	for {
		rssItems = append(rssItems, rssFile.RssItems...)
		if len(rssFile.NextUrl) == 0 {
			log.Println("There's no next URL so fetching RSS XMLs is over.")
			break
		} else {
			log.Println("Fetching another RSS XML:", rssFile.NextUrl)
		}

		rssFile, err = deviantart.FetchRssFile(rssFile.NextUrl)
		if err != nil {
			return
		}
	}
	return
}

// A single saved deviation
type SavedDeviation struct {
	RssItem  deviantart.RssItem
	Filename string
}

// One full fetch, all deviations, their saved filenames etc.
type DeviantFetch struct {
	SavedDeviations []SavedDeviation
	Timestamp       time.Time
}

// Generate UUID string
// Grabbed from https://play.golang.org/p/4FkNSiUDMg
func newUuid() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		// TODO Add logging
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func downloadImages(dirname, url string, dryRun bool) string {
	uuid, _ := newUuid()
	pieces := strings.Split(url, "/")
	fpath := filepath.Join(dirname, uuid, pieces[len(pieces)-1])
	if dryRun {
		return fpath
	} else {
		dirpath := filepath.Join(dirname, uuid)
		err := os.MkdirAll(dirpath, 0700)
		if err != nil {
			// TODO Return err
			// TODO Add logging
			fmt.Fprintln(os.Stderr, "Failed to create directories for", dirpath)
			fmt.Fprintln(os.Stderr, err)
			return ""
		}
	}

	src, err := http.Get(url)
	if err != nil {
		// TODO Return err
		// TODO Add logging
		fmt.Fprintln(os.Stderr, "Failed to fetch image.")
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	defer src.Body.Close()

	dest, err := os.Create(fpath)
	if err != nil {
		// TODO Return err
		// TODO Add logging
		fmt.Fprintln(os.Stderr, "Failed to create image file:", fpath)
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	defer dest.Close()
	defer fmt.Println("Deviation downloaded:", fpath)

	byteCount, err := io.Copy(dest, src.Body)
	if err != nil {
		// TODO Return err
		// TODO Add logging
		fmt.Fprintln(os.Stderr, "Failed to copy image to file from", fpath)
		fmt.Fprintln(os.Stderr, "Count of bytes copied:", byteCount)
		fmt.Fprintln(os.Stderr, err)
		return ""
	}

	return fpath
}

func toDeviantFetch(rssItems []deviantart.RssItem, dirname string) DeviantFetch {
	var savedDeviations []SavedDeviation
	for _, eachItem := range rssItems {
		aSavedDeviation := SavedDeviation{
			RssItem:  eachItem,
			Filename: downloadImages(dirname, eachItem.Url, false),
		}
		savedDeviations = append(savedDeviations, aSavedDeviation)
	}
	return DeviantFetch{
		SavedDeviations: savedDeviations,
		Timestamp:       time.Now(),
	}
}

func app(username, dirpath string) (err error) {
	var rssItems []deviantart.RssItem
	rssItems, err = fetchItems(username)
	if err != nil {
		return
	}
	deviantFetch := toDeviantFetch(rssItems, dirpath)
	jsonBytes, err := json.Marshal(deviantFetch)
	if err != nil {
		// TODO Add logging
		fmt.Fprintln(os.Stderr, "Conversion to json failed.")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	err = ioutil.WriteFile("deviantFetch.json", jsonBytes, 0644)
	if err != nil {
		// TODO Add logging
		fmt.Fprintln(os.Stderr, err)
		return
	}

	return
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Missing username")
		fmt.Printf("Usage: %s {username}\n", os.Args[0])
		return
	}

	username := strings.TrimSpace(os.Args[1])
	if len(username) == 0 {
		fmt.Println("Username can't be empty")
		os.Exit(1)
	}

	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create a temporary directory.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Println("Create directory:", dirpath)
	err = app(username, dirpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}
