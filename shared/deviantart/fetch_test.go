package deviantart

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"golang.org/x/net/html"
)

func TestExtractDownloadLinkURL(t *testing.T) {
	run := func(name, link, expected string) {
		t.Run(name, func(t *testing.T) {
			// SETUP SUT
			reader := strings.NewReader(link)
			tokenizer := html.NewTokenizer(reader)
			tokenizer.Next()

			// EXERCISE
			actual := extractDownloadLinkURL(tokenizer)

			// VERIFY
			testza.AssertEqual(t, expected, actual)
		})
	}

	run(
		"Class first",
		`<a class="dev-page-download" href="classFirst"/>`,
		"classFirst")
	run(
		"Href first",
		`<a href="hrefFirst" class=" dev-page-download"/>`,
		"hrefFirst")
	run(
		"Just a",
		"<a/>",
		"")
	run(
		"Wrong class",
		`<a href="wontBeReturned" class="not-the-right-one"`,
		"")
}

func TestExtractDownloadURL(t *testing.T) {
	run := func(name, response, expected string) {
		t.Run(name, func(t *testing.T) {
			testza.AssertEqual(
				t,
				expected,
				ExtractDownloadURL(strings.NewReader(response)))
		})
	}

	run(
		"Just a inside div",
		`<div><a href="hellgod" class="dev-page-download"></a></div>`,
		"hellgod")

	href := "https://realistic.com/real.jpg?funny=no"
	run(
		"More realistic test",
		fmt.Sprintf(
			`<html>
				<head>
					<title>Godlike Creation</title>
				</head>
				<body>
					<div class="dev-page-download">
						<a
							class="dev-page-download"
							href="%s">
					</div>
				</body>
			</html>`,
			href),
		href)
}

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
