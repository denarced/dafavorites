package deviantart

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"golang.org/x/net/html"
)

func TestExtractDownloadLinkURL(t *testing.T) {
	test := func(link, expected string) {
		// SETUP SUT
		reader := strings.NewReader(link)
		tokenizer := html.NewTokenizer(reader)
		tokenizer.Next()

		// EXERCISE
		actual := extractDownloadLinkURL(tokenizer)

		// VERIFY
		if expected != actual {
			t.Errorf("Expected '%s' but got '%s'", expected, actual)
		}
	}

	t.Run("Class first", func(t *testing.T) {
		test(`<a class="dev-page-download" href="classFirst"/>`, "classFirst")
	})
	t.Run("Href first", func(t *testing.T) {
		test(`<a href="hrefFirst" class=" dev-page-download"/>`, "hrefFirst")
	})
	t.Run("Just a", func(t *testing.T) {
		test("<a/>", "")
	})
	t.Run("Wrong class", func(t *testing.T) {
		test(`<a href="wontBeReturned" class="not-the-right-one"`, "")
	})
}

func TestExtractDownloadURL(t *testing.T) {
	test := func(response, expected string) {
		// EXERCISE
		actual := ExtractDownloadURL(strings.NewReader(response))

		// VERIFY
		if expected != actual {
			t.Errorf("Expected '%s' but got '%s'", expected, actual)
		}
	}

	t.Run("Just a inside div", func(t *testing.T) {
		test(
			`<div><a href="hellgod" class="dev-page-download"></a></div>`,
			"hellgod")
	})
	t.Run("More realistic test", func(t *testing.T) {
		href := "https://realistic.com/real.jpg?funny=no"
		htmlTemplate := `
			<html>
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
			</html>
		`
		html := fmt.Sprintf(htmlTemplate, href)
		test(html, href)
	})
}

func TestToRssFile(t *testing.T) {
	// SETUP SUT
	rssBytes, err := ioutil.ReadFile("testdata/rss.xml")
	if err != nil {
		t.Errorf("Failed to read test file rss.xml: %s", err)
	}

	// EXERCISE
	rssFile, err := ToRssFile(bytes.NewReader(rssBytes))

	// VERIFY
	if err != nil {
		t.Errorf("Error received from ToRssFile: %v", err)
	}
	if len(rssFile.RssItems) != 60 {
		t.Errorf("Expected %d items but got %d", 60, len(rssFile.RssItems))
	}
	expectedNextURL := "https://backend.deviantart.com/rss.xml?" +
		"type=deviation&offset=120&q=favedbyid%3A4471416&order=9"
	if expectedNextURL != rssFile.NextURL {
		t.Errorf(
			"Next URL isn't as expected:\n%s",
			diff.LineDiff(expectedNextURL, rssFile.NextURL))
	}

	expectedFirstItem := RssItem{
		Title:           "Leya",
		Link:            "https://art0fck.deviantart.com/art/Leya-671530106",
		GUID:            "https://art0fck.deviantart.com/art/Leya-671530106",
		PublicationDate: "Tue, 28 Mar 2017 03:37:53 PDT",
		Author:          "art0fCK",
		URL:             "https://pre00.deviantart.net/04fc/th/pre/f/2017/087/d/3/d3cf26870151df8b05491ec8c1242fc8-db3t7y2.jpg",
		Dimensions: Dimensions{
			Width:  730,
			Height: 1095,
		},
	}
	actualFirstItem := rssFile.RssItems[0]
	if expectedFirstItem != actualFirstItem {
		t.Errorf(
			"Mismatch with first RSS item:\n%s",
			diff.LineDiff(
				fmt.Sprintf("%v", expectedFirstItem),
				fmt.Sprintf("%v", actualFirstItem)))
	}

	expectedLastItem := RssItem{
		Title:           "double fluo",
		Link:            "https://abrito.deviantart.com/art/double-fluo-64794797",
		GUID:            "https://abrito.deviantart.com/art/double-fluo-64794797",
		PublicationDate: "Thu, 13 Sep 2007 07:48:45 PDT",
		Author:          "ABrito",
		URL:             "https://orig00.deviantart.net/8878/f/2007/256/0/9/no_title_33_by_abrito.jpg",
		Dimensions: Dimensions{
			Width:  554,
			Height: 750,
		},
	}
	actualLastItem := rssFile.RssItems[len(rssFile.RssItems)-1]
	if expectedLastItem != actualLastItem {
		t.Errorf(
			"Mismatch with last RSS item:\n%s",
			diff.LineDiff(
				fmt.Sprintf("%v", expectedLastItem),
				fmt.Sprintf("%v", actualLastItem)))
	}
}
