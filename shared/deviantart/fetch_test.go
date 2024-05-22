package deviantart

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestToRssFile(t *testing.T) {
	// SETUP SUT
	rssBytes, err := ioutil.ReadFile("testdata/rss.xml")
	testza.AssertNil(t, err)

	// EXERCISE
	rssFile, err := ToRssFile(bytes.NewReader(rssBytes))

	// VERIFY
	testza.AssertNil(t, err, "Error received from ToRssFile.")
	testza.AssertEqual(
		t,
		60,
		len(rssFile.RssItems), "Unexpected count of RssItems.")
	expectedNextURL := "https://backend.deviantart.com/rss.xml?" +
		"type=deviation&offset=120&q=favedbyid%3A4471416&order=9"
	testza.AssertEqual(
		t,
		expectedNextURL,
		rssFile.NextURL,
		"Unexpected NextURL.")

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
	testza.AssertEqual(
		t,
		expectedFirstItem,
		actualFirstItem,
		"Mismatched first RSS item.")

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
	testza.AssertEqual(
		t,
		expectedLastItem,
		actualLastItem,
		"Mismatched last RSS item.")
}
