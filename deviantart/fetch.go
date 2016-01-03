package deviantart

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type RssItem struct {
	Title  string
	Link   string
	Author string
	Url    string
	Width  int
	Height int
}

type RssFile struct {
	NextUrl  string
	RssItems []RssItem
}

type LinkElement struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type ItemCreditElement struct {
	Role  string `xml:"role,attr"`
	Value string `xml:",chardata"`
}

type ItemContentElement struct {
	Url    string `xml:"url,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type RssItemElement struct {
	Title   string              `xml:"title"`
	Link    string              `xml:"link"`
	Url     string              `xml:"url"`
	Width   int                 `xml:"width"`
	Height  int                 `xml:"height"`
	Credits []ItemCreditElement `xml:"credit"`
	Content ItemContentElement  `xml:"content"`
}

type ChannelElement struct {
	Links    []LinkElement    `xml:"link"`
	RssItems []RssItemElement `xml:"item"`
}

type RssElement struct {
	XMLName xml.Name       `xml:"rss"`
	Channel ChannelElement `xml:"channel"`
}

func itemElementsToItems(elements []RssItemElement) []RssItem {
	var rssItems []RssItem
	for _, each := range elements {
		var author string
		for _, eachCredit := range each.Credits {
			if eachCredit.Role == "author" && strings.HasPrefix(eachCredit.Value, "http") == false {
				author = eachCredit.Value
				break
			}
		}
		rssItems = append(rssItems, RssItem{
			Title:  each.Title,
			Link:   each.Link,
			Author: author,
			Url:    each.Content.Url,
			Width:  each.Content.Width,
			Height: each.Content.Height,
		})
	}
	return rssItems
}

func FetchRssFile(url string) RssFile {
	log.Printf("Fetch rss file %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch rss file: %v\n", err)
		return RssFile{}
	}
	defer resp.Body.Close()
	contentBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read fetched rss file: %v\n", err)
		return RssFile{}
	}

	rssElement := RssElement{}
	xml.Unmarshal(contentBytes, &rssElement)
	var rssItems []RssItem = itemElementsToItems(rssElement.Channel.RssItems)

	var next string
	for _, each := range rssElement.Channel.Links {
		if each.Rel == "next" {
			next = each.Href
			break
		}
	}

	return RssFile{
		NextUrl:  next,
		RssItems: rssItems,
	}
}
