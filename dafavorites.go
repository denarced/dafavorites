package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/denarced/dafavorites/deviantart"
)

const baseRss = "http://backend.deviantart.com/rss.xml?q=favby%3A___usern___&type=deviation"
const logFlags = log.LstdFlags | log.Lshortfile

var (
	infoLogger  = log.New(os.Stdout, "INFO ", logFlags)
	errorLogger = log.New(os.Stderr, "ERROR ", logFlags)
)

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
		errorLogger.Println("UUID generation failed:", err)
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	endResult := fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:])
	return endResult, nil
}

type DownloadParams struct {
	Client  *http.Client
	Dirname string
	Url     string
	DryRun  bool
	Uuid    string
	Prefix  string
}

// Download file url to a directory under dirname unless dryRun is true. The
// directory under dirname is created only for the downloaded file and its name
// TODO Fix
// is a UUID. An example call:
//     downloadImages("/tmp/deviations", "http://site.com/image.jpg", false)
//     "/tmp/deviations/06c6e05e-e22a-43d2-9e69-e198825e07fd/image.jpg"
func downloadImages(params DownloadParams) string {
	filename := deriveFilename(params.Prefix, params.Url)
	fpath := filepath.Join(params.Dirname, params.Uuid, filename)
	if params.DryRun {
		infoLogger.Println("Dry run: skip download of ", fpath)
		return fpath
	} else {
		dirpath := filepath.Join(params.Dirname, params.Uuid)
		err := os.MkdirAll(dirpath, 0700)
		if err != nil {
			errorLogger.Printf("Failed to create path. Path: %s. Error: %v.\n", dirpath, err)
			return ""
		}
	}

	src, err := params.Client.Get(params.Url)
	if err != nil {
		errorLogger.Println("Failed to fetch image:", err)
		return ""
	}
	defer src.Body.Close()

	dest, err := os.Create(fpath)
	if err != nil {
		errorLogger.Printf("Failed to create image file. Filepath: %v. Error: %v.\n", fpath, err)
		return ""
	}
	defer dest.Close()
	defer infoLogger.Println("Deviation downloaded:", fpath)

	byteCount, err := io.Copy(dest, src.Body)
	if err != nil {
		errorLogger.Println("Failed to copy image to file from", fpath)
		errorLogger.Println("Count of bytes copied:", byteCount)
		errorLogger.Println("Error:", err)
		return ""
	}

	return fpath
}

func deriveFilename(prefix, url string) string {
	pieces := strings.Split(url, "/")
	// e.g. image.jpg?token=blaablaa or
	//      image.jpg
	withExtra := pieces[len(pieces)-1]
	// e.g. [image.jpg token=blaablaa] or
	//      [image.jpg]
	extraPieces := strings.Split(withExtra, "?")
	return prefix + "_" + extraPieces[0]
}

// Fetch RSS files and pass the deviations to be downloaded. The RSSs are
// fetched for user username and each deviation is passed to rssItemChan. Once
// done, the channel finished is closed to signal that work is done.
func fetchRss(username string, rssItemChan chan deviantart.RssItem, finished chan struct{}) {
	defer close(finished)

	url := strings.Replace(baseRss, "___usern___", username, 1)
	rssFile, err := deviantart.FetchRssFile(url)
	if err != nil {
		errorLogger.Println("Failed to fetch RSS file:", err)
		return
	}
	for {
		// Pass favorite deviations to be downloaded
		for _, each := range rssFile.RssItems {
			rssItemChan <- each
		}

		// Fetch more deviations if there are some
		if len(rssFile.NextUrl) == 0 {
			break
		}

		rssFile, err = deviantart.FetchRssFile(rssFile.NextUrl)
		if err != nil {
			errorLogger.Println("Failed to fetch RSS file:", err)
			return
		}
	}
}

// Download and save deviations. Jobs are received from rssItemChan and results
// are passed to savedDeviationChan. Parameter id is the identifier and it isn't
// functional. It'll be used merely in any logging or printouts. Once the
// channel rssItemChan no longer provides jobs to perform, waitGroup.Done() is
// called in order to inform the caller that this method has completed. If
// dryRun is true, nothing is really downloaded but otherwise the process is
// executed in a normal fashion.
func saveDeviations(id int, dirpath string, rssItemChan chan deviantart.RssItem, savedDeviationChan chan SavedDeviation, waitGroup *sync.WaitGroup, dryRun bool) {
	defer waitGroup.Done()

	infoLogger.Println("Starting download worker", id)
	for each := range rssItemChan {
		infoLogger.Printf("Worker %d about to start downloading %s\n", id, each.Url)
		cookieJar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: cookieJar,
		}
		uuid, err := newUuid()
		if err != nil {
			errorLogger.Printf(
				"UUID generation error when working with %s: %v\n",
				each.Url,
				err)
			continue
		}
		params := DownloadParams{
			Client:  client,
			Dirname: dirpath,
			Url:     each.Url,
			DryRun:  dryRun,
			Uuid:    uuid,
			Prefix:  "regular",
		}
		filepath := downloadImages(params)
		if len(filepath) == 0 {
			// Nothing to be done if the download failed as the error should
			// have been reported by the called function.
			continue
		}
		savedDeviationChan <- SavedDeviation{
			RssItem:  each,
			Filename: filepath,
		}

		if dryRun {
			continue
		}

		dlUrl := deviantart.ExtractDownloadUrl(client, each.Link)
		if dlUrl == "" {
			continue
		}

		dlParams := params
		dlParams.Url = dlUrl
		dlParams.Prefix = "large"
		filepath = downloadImages(dlParams)
		each.Url = dlUrl
		dimensions := extractDimensions(filepath)
		each.Dimensions = dimensions

		savedDeviationChan <- SavedDeviation{
			RssItem:  each,
			Filename: filepath,
		}
	}

	infoLogger.Println("Quitting download worker", id)
}

func extractDimensions(filepath string) (dimensions deviantart.Dimensions) {
	reader, err := os.Open(filepath)
	if err != nil {
		errorLogger.Printf(
			"Error opening image (%s): %s. Leaving dimensions to zeros.",
			filepath,
			err)
		return
	}

	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		errorLogger.Printf(
			"Error decoding image (%s): %s\n",
			filepath,
			err)
		return
	}

	dimensions.Width = img.Width
	dimensions.Height = img.Height
	return
}

// Collected downloaded deviations into a single DeviantFetch. The deviations
// are received from savedDeviationChan and the end result is passed to
// deviantFetchChan.
func collectSavedDeviations(savedDeviationChan chan SavedDeviation, deviantFetchChan chan DeviantFetch) {
	var deviations []SavedDeviation
	for each := range savedDeviationChan {
		infoLogger.Println("Deviation has arrived to be collected:", each.Filename)
		deviations = append(deviations, each)
	}
	deviantFetchChan <- DeviantFetch{
		SavedDeviations: deviations,
		Timestamp:       time.Now(),
	}
}

// Fetch user username's favorite deviations to directory dirpath. Several
// images can be downloaded in parallel according to dlWorkerCount. It's value
// must be at least 1. Return information on all fetched deviations.
func fetchFavorites(username, dirpath string, dlWorkerCount int) DeviantFetch {
	// Buffered channel so that fetching RSSs isn't completely blocked by
	// downloaders.
	rssItemChan := make(chan deviantart.RssItem, 500)
	rssFinished := make(chan struct{})
	go fetchRss(username, rssItemChan, rssFinished)

	dlWaitGroup := sync.WaitGroup{}
	savedDeviationChan := make(chan SavedDeviation)
	for i := 0; i < dlWorkerCount; i++ {
		dlWaitGroup.Add(1)
		go saveDeviations(i, dirpath, rssItemChan, savedDeviationChan, &dlWaitGroup, false)
	}

	deviantFetchChan := make(chan DeviantFetch)
	go collectSavedDeviations(savedDeviationChan, deviantFetchChan)

	// Wait until RSS downloads have finished
	<-rssFinished
	infoLogger.Println("Go routine for fetching RSS files has finished.")
	// Close RSS channel in order to signal to downloaders that there's no more
	// jobs coming.
	close(rssItemChan)
	// Wait for the downloaders to finish
	dlWaitGroup.Wait()
	infoLogger.Println("All downloaders have finished.")
	// Downloaders finished so close chan so that collector stops waiting
	close(savedDeviationChan)
	// And finally get information on all favorite deviations from collector
	return <-deviantFetchChan
}

// Save information on fetched deviations to file filename.
func saveJson(deviantFetch DeviantFetch, filename string) error {
	jsonBytes, err := json.Marshal(deviantFetch)
	if err != nil {
		errorLogger.Println("Conversion to json failed:", err)
		return err
	}

	err = ioutil.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		errorLogger.Println("Error writing JSON:", err)
		return err
	}

	return nil
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

	fmt.Println("Create temporary directory")
	dirpath, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create a temporary directory.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	deviantFetch := fetchFavorites(username, dirpath, 4)
	infoLogger.Println("Deviations fetched.")
	err = saveJson(deviantFetch, "deviantFetch.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}
