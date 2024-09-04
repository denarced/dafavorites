package deviantart

import (
	"bytes"
	"os"
	"testing"

	djson "github.com/denarced/dafavorites/shared/deviantart/json"
	"github.com/stretchr/testify/assert"
)

func TestToRssFile(t *testing.T) {
	// SETUP SUT
	rssBytes, err := os.ReadFile("testdata/rss.xml")
	req := assert.New(t)
	req.Nil(err)

	// EXERCISE
	rssFile, err := toRssFile(bytes.NewReader(rssBytes))

	// VERIFY
	req.Nil(err, "Error received from toRssFile.")
	req.Equal(60, len(rssFile.rssItems), "Unexpected count of RssItems.")
	expectedNextURL := "https://backend.deviantart.com/rss.xml?" +
		"type=deviation&offset=120&q=favedbyid%3A4471416&order=9"
	req.Equal(expectedNextURL, rssFile.nextURL, "Unexpected NextURL.")

	expectedFirstItem := djson.RssItem{
		Title:           "Leya",
		Link:            "https://art0fck.deviantart.com/art/Leya-671530106",
		GUID:            "https://art0fck.deviantart.com/art/Leya-671530106",
		PublicationDate: "Tue, 28 Mar 2017 03:37:53 PDT",
		Author:          "art0fCK",
		URL: "https://pre00.deviantart.net/04fc/th/pre/f/2017/087/" +
			"d/3/d3cf26870151df8b05491ec8c1242fc8-db3t7y2.jpg",
		Dimensions: djson.Dimensions{
			Width:  730,
			Height: 1095,
		},
	}
	actualFirstItem := rssFile.rssItems[0]
	req.Equal(expectedFirstItem, actualFirstItem, "Mismatched first RSS item.")

	expectedLastItem := djson.RssItem{
		Title: "double fluo",
		Link: "https://abrito.deviantart.com/art/" +
			"double-fluo-64794797",
		GUID: "https://abrito.deviantart.com/art/" +
			"double-fluo-64794797",
		PublicationDate: "Thu, 13 Sep 2007 07:48:45 PDT",
		Author:          "ABrito",
		URL: "https://orig00.deviantart.net/" +
			"8878/f/2007/256/0/9/no_title_33_by_abrito.jpg",
		Dimensions: djson.Dimensions{
			Width:  554,
			Height: 750,
		},
	}
	actualLastItem := rssFile.rssItems[len(rssFile.rssItems)-1]
	req.Equal(expectedLastItem, actualLastItem, "Mismatched last RSS item.")
}

func TestDeriveFilename(t *testing.T) {
	run := func(name, prefix, url, expected string) {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, expected, deriveFilename(prefix, url))
		})
	}

	run("No prefix", "", "http://a.com/me.jpg", "me.jpg")
	run(
		"With URL parameters",
		"dox",
		"http://a.com/me.jpg?param=value", "dox_me.jpg")
	run(
		"With dirs in URL",
		"longer",
		"http://b.com/shut/down/more.jpg",
		"longer_more.jpg")
}
