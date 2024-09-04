package deviantart

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToRssFile(t *testing.T) {
	// SETUP SUT
	rssBytes, err := os.ReadFile("testdata/rss.xml")
	req := assert.New(t)
	req.Nil(err)

	// EXERCISE
	rssFile, err := ToRssFile(bytes.NewReader(rssBytes))

	// VERIFY
	req.Nil(err, "Error received from ToRssFile.")
	req.Equal(60, len(rssFile.RssItems), "Unexpected count of RssItems.")
	expectedNextURL := "https://backend.deviantart.com/rss.xml?" +
		"type=deviation&offset=120&q=favedbyid%3A4471416&order=9"
	req.Equal(expectedNextURL, rssFile.NextURL, "Unexpected NextURL.")

	expectedFirstItem := RssItem{
		Title:           "Leya",
		Link:            "https://art0fck.deviantart.com/art/Leya-671530106",
		GUID:            "https://art0fck.deviantart.com/art/Leya-671530106",
		PublicationDate: "Tue, 28 Mar 2017 03:37:53 PDT",
		Author:          "art0fCK",
		URL: "https://pre00.deviantart.net/04fc/th/pre/f/2017/087/" +
			"d/3/d3cf26870151df8b05491ec8c1242fc8-db3t7y2.jpg",
		Dimensions: Dimensions{
			Width:  730,
			Height: 1095,
		},
	}
	actualFirstItem := rssFile.RssItems[0]
	req.Equal(expectedFirstItem, actualFirstItem, "Mismatched first RSS item.")

	expectedLastItem := RssItem{
		Title: "double fluo",
		Link: "https://abrito.deviantart.com/art/" +
			"double-fluo-64794797",
		GUID: "https://abrito.deviantart.com/art/" +
			"double-fluo-64794797",
		PublicationDate: "Thu, 13 Sep 2007 07:48:45 PDT",
		Author:          "ABrito",
		URL: "https://orig00.deviantart.net/" +
			"8878/f/2007/256/0/9/no_title_33_by_abrito.jpg",
		Dimensions: Dimensions{
			Width:  554,
			Height: 750,
		},
	}
	actualLastItem := rssFile.RssItems[len(rssFile.RssItems)-1]
	req.Equal(expectedLastItem, actualLastItem, "Mismatched last RSS item.")
}
