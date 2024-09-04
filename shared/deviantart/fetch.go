// Package deviantart contains all code strictly not related CLI.
package deviantart

import (
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	djson "github.com/denarced/dafavorites/shared/deviantart/json"
	dxml "github.com/denarced/dafavorites/shared/deviantart/xml"
	"github.com/denarced/dafavorites/shared/shared"
)

const (
	baseRss = "http://backend.deviantart.com/rss.xml" +
		"?q=favby%3A___usern___&type=deviation"
)

// RssFile is the items of the one Deviant Art RSS file and the next one's URL
type rssFile struct {
	nextURL  string
	rssItems []djson.RssItem
}

// Convert deviant art structures to our own
func itemElementsToItems(elements []dxml.RssItemElement) []djson.RssItem {
	if len(elements) == 0 {
		return []djson.RssItem{}
	}

	rssItems := make([]djson.RssItem, 0, len(elements))
	for _, each := range elements {
		rssItems = append(
			rssItems,
			djson.RssItem{
				Title:           each.Title,
				Link:            each.Link,
				GUID:            each.GUID,
				PublicationDate: each.PublicationDate,
				Author:          extractAuthor(each.Credits),
				URL:             each.Content.URL,
				Dimensions: djson.Dimensions{
					Width:  each.Content.Width,
					Height: each.Content.Height}})
	}
	return rssItems
}

func extractAuthor(credits []dxml.ItemCreditElement) string {
	for _, eachCredit := range credits {
		if eachCredit.Role == "author" &&
			!strings.HasPrefix(eachCredit.Value, "http") {
			return eachCredit.Value
		}
	}
	return ""
}

// ToRssFile converts reader contents to an rssFile
func toRssFile(reader io.Reader) (rssFile, error) {
	contentBytes, err := io.ReadAll(reader)
	if err != nil {
		shared.Logger.Error("Failed to read fetched rss file.", "error", err)
		return rssFile{}, err
	}

	rssElement := dxml.RssElement{}
	if err = xml.Unmarshal(contentBytes, &rssElement); err != nil {
		shared.Logger.Error("Failed to unmarshal XML.", "error", err)
		return rssFile{}, err
	}
	rssItems := itemElementsToItems(rssElement.Channel.RssItems)

	return rssFile{
		nextURL:  extractNextHref(rssElement.Channel.Links),
		rssItems: rssItems,
	}, nil
}

func extractNextHref(links []dxml.LinkElement) string {
	for _, each := range links {
		if each.Rel == "next" {
			return each.Href
		}
	}
	return ""
}

// NewUUID generates a single UUID string
// Grabbed from https://play.golang.org/p/4FkNSiUDMg
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		shared.Logger.Error("UUID generation failed.", "error", err)
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

// DownloadParams for downloading images from deviant art.
type downloadParams struct {
	// Client to use to download the image. Must not be null.
	client *http.Client
	// Dirname is the root dir into which images are downloaded.
	dirname string
	// URL for the image to download.
	url string
	// Don't actually download anything when true.
	dryRun bool
	// UUID to act as a sub dir under Dirname.
	uuid string
	// Filename for the image.
	filename string
}

// Download file params.url with params as a specification.
// Return the downloaded file's filepath.
func downloadImages(params downloadParams) string {
	fpath := filepath.Join(params.dirname, params.uuid, params.filename)
	if params.dryRun {
		shared.Logger.Debug("Dry run: skip download.", "filepath", fpath)
		return ""
	}
	dirpath := filepath.Join(params.dirname, params.uuid)
	if err := os.MkdirAll(dirpath, 0700); err != nil {
		shared.Logger.Error("Failed to create path.", "dirpath", dirpath, "error", err)
		return ""
	}

	src, err := params.client.Get(params.url)
	if err != nil {
		shared.Logger.Error("Failed to fetch image.", "error", err)
		return ""
	}
	defer src.Body.Close()

	imageBytes, err := io.ReadAll(src.Body)
	imageSize := int64(len(imageBytes))
	shared.Logger.Debug("Fetched image.", "filepath", fpath, "size", imageSize)
	if imageSize <= 0 {
		return ""
	}

	dest, err := os.Create(fpath)
	if err != nil {
		shared.Logger.Error("Failed to create image file.", "filepath", fpath, "error", err)
		return ""
	}
	defer dest.Close()
	defer shared.Logger.Debug("Deviation downloaded.", "filepath", fpath)

	byteCount, err := dest.Write(imageBytes)
	if err != nil {
		shared.Logger.Error(
			"Failed to copy image to file.",
			"filepath",
			fpath,
			"byte count",
			byteCount,
			"error",
			err,
		)
		return ""
	}

	return fpath
}

func deriveFilename(prefix, url string) string {
	pieces := strings.Split(url, "/")
	// E.g. image.jpg?token=blaablaa or
	//      image.jpg
	withExtra := pieces[len(pieces)-1]
	// E.g. [image.jpg token=blaablaa] or
	//      [image.jpg]
	extraPieces := strings.Split(withExtra, "?")
	separator := "_"
	if prefix == "" {
		separator = ""
	}
	return prefix + separator + extraPieces[0]
}

func fetchAndReadRss(url string) (rssFile, error) {
	resp, err := fetchRssFile(url)
	if err != nil {
		return rssFile{}, err
	}
	defer resp.Body.Close()
	return toRssFile(resp.Body)
}

// Fetch RSS files and pass the deviations to be downloaded. The RSSs are
// fetched for user username and each deviation is passed to rssItemChan. Once
// done, the channel finished is closed to signal that work is done.
func fetchRss(
	username string,
	rssItemChan chan djson.RssItem,
	finished chan struct{}) {
	defer close(finished)

	url := strings.Replace(baseRss, "___usern___", username, 1)
	rssFile, err := fetchAndReadRss(url)
	if err != nil {
		return
	}
	for {
		// Pass favorite deviations to be downloaded
		for _, each := range rssFile.rssItems {
			rssItemChan <- each
		}
		// Fetch more deviations if there are some
		if len(rssFile.nextURL) == 0 {
			break
		}

		rssFile, err = fetchAndReadRss(rssFile.nextURL)
		if err != nil {
			return
		}
	}
}

func fetchRssFile(url string) (resp *http.Response, err error) {
	shared.Logger.Debug("About to fetch RSS file.", "url", url)
	resp, err = http.Get(url)
	if err != nil {
		shared.Logger.Error("Failed to fetch RSS file.", "error", err)
	}
	shared.Logger.Info("RSS file fetched.", "url", url)
	return
}

// Download and save deviations. Jobs are received from rssItemChan and results
// are passed to savedDeviationChan. Parameter id is the identifier and it isn't
// functional. It'll be used merely in any logging or printouts. Once the
// channel rssItemChan no longer provides jobs to perform, waitGroup.Done() is
// called in order to inform the caller that this method has completed. If
// dryRun is true, nothing is really downloaded but otherwise the process is
// executed in a normal fashion.
func saveDeviations(
	id int,
	dirpath string,
	rssItemChan chan djson.RssItem,
	savedDeviationChan chan djson.SavedDeviation,
	waitGroup *sync.WaitGroup,
	dryRun bool,
) {
	defer waitGroup.Done()

	shared.Logger.Debug("Starting download worker.", "ID", id)
	for each := range rssItemChan {
		shared.Logger.Debug(
			"Worker about to start downloading.",
			"id",
			id,
			"url",
			each.URL)
		shared.Logger.Debug("Worker: create cookie jar.", "id", id)
		cookieJar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar: cookieJar,
		}
		shared.Logger.Debug("Worker: create UUID.", "id", id)
		uuid, err := newUUID()
		if err != nil {
			shared.Logger.Error("UUID generation failed.", "url", each.URL, "error", err)
			continue
		}
		filename := deriveFilename("", each.URL)
		params := downloadParams{
			client:   client,
			dirname:  dirpath,
			url:      each.URL,
			dryRun:   dryRun,
			uuid:     uuid,
			filename: filename,
		}
		shared.Logger.Debug("Worker: download image.", "id", id, "url", params.url)
		filep := downloadImages(params)
		if len(filep) == 0 {
			// Nothing to be done if the download failed as the error should
			// have been reported by the called function.
			continue
		}
		savedDeviationChan <- djson.SavedDeviation{
			RssItem:  each,
			Filename: filep,
		}
	}

	shared.Logger.Info("Quitting download worker.", "id", id)
}

func deriveURL(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	return res.Request.URL.String(), nil
}

// Collected downloaded deviations into a single DeviantFetch. The deviations
// are received from savedDeviationChan and the end result is passed to
// deviantFetchChan.
func collectSavedDeviations(
	savedDeviationChan chan djson.SavedDeviation,
	deviantFetchChan chan djson.DeviantFetch,
) {
	var deviations []djson.SavedDeviation
	for each := range savedDeviationChan {
		shared.Logger.Info("Deviation has arrived to be collected.", "filename", each.Filename)
		deviations = append(deviations, each)
	}
	deviantFetchChan <- djson.DeviantFetch{
		SavedDeviations: deviations,
		Timestamp:       time.Now(),
	}
}

// FetchFavorites fetches user username's favorite deviations to directory dirpath. Several
// images can be downloaded in parallel according to dlWorkerCount. It's value
// must be at least 1. Return information on all fetched deviations.
func FetchFavorites(username, dirpath string, dlWorkerCount int) djson.DeviantFetch {
	// Buffered channel so that fetching RSSs isn't completely blocked by
	// downloaders.
	rssItemChan := make(chan djson.RssItem, 500)
	rssFinished := make(chan struct{})
	go fetchRss(username, rssItemChan, rssFinished)

	dlWaitGroup := sync.WaitGroup{}
	savedDeviationChan := make(chan djson.SavedDeviation)
	for i := 0; i < dlWorkerCount; i++ {
		dlWaitGroup.Add(1)
		go saveDeviations(
			i,
			dirpath,
			rssItemChan,
			savedDeviationChan,
			&dlWaitGroup,
			false)
	}

	deviantFetchChan := make(chan djson.DeviantFetch)
	go collectSavedDeviations(savedDeviationChan, deviantFetchChan)

	// Wait until RSS downloads have finished
	<-rssFinished
	shared.Logger.Info("Go routine for fetching RSS files has finished.")
	// Close RSS channel in order to signal to downloaders that there's no more
	// jobs coming.
	close(rssItemChan)
	// Wait for the downloaders to finish
	dlWaitGroup.Wait()
	shared.Logger.Info("All downloaders have finished.")
	// Downloaders finished so close chan so that collector stops waiting
	close(savedDeviationChan)
	// And finally get information on all favorite deviations from collector
	return <-deviantFetchChan
}

// SaveJSON saves information on fetched deviations to file filename.
func SaveJSON(deviantFetch djson.DeviantFetch, filename string) error {
	jsonBytes, err := json.Marshal(deviantFetch)
	if err != nil {
		shared.Logger.Error("Conversion to json failed.", "error", err)
		return err
	}

	err = os.WriteFile(filename, jsonBytes, 0644)
	if err != nil {
		shared.Logger.Error("Error writing JSON.", "error", err)
		return err
	}

	return nil
}
