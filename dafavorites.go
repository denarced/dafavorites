package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/denarced/dafavorites/deviantart"
)

const baseRss = "http://backend.deviantart.com/rss.xml?q=favby%3A___usern___%2F2573873&type=deviation"

func fetchItems(username string) []deviantart.RssItem {
	log.Printf("Fetch information on %s's favorite deviations", username)
	// Fetch first RssFile
	// Grab all items and store them
	// If RssFile.NextUrl contains something, continue with it
	var rssFile deviantart.RssFile
	url := strings.Replace(baseRss, "___usern___", username, 1)
	rssFile = deviantart.FetchRssFile(url)
	var rssItems []deviantart.RssItem
	for {
		rssItems = append(rssItems, rssFile.RssItems...)
		if len(rssFile.NextUrl) == 0 {
			break
		}

		rssFile = deviantart.FetchRssFile(rssFile.NextUrl)
	}
	return rssItems
}

func downloadImages(rssItems []deviantart.RssItem) {
	jsonBytes, err := json.Marshal(rssItems)
	if err == nil {
		outFilen := "deviations.json"
		err = ioutil.WriteFile(outFilen, jsonBytes, 0644)
		if err != nil {
			log.Printf("Failed to write favorites to %s. Error: %v\n", outFilen, err)
			fmt.Printf("%s\n", jsonBytes)
		} else {
			log.Printf("Favorites successfully written to %s\n", outFilen)
		}
	} else {
		fmt.Printf("Error: %s\n", err)
	}
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Missing username")
		fmt.Printf("Usage: %s {username}\n", os.Args[0])
	} else {
		username := strings.TrimSpace(os.Args[1])
		if len(username) == 0 {
			fmt.Println("Username can't be empty")
		} else {
			downloadImages(fetchItems(username))
		}
	}
}
