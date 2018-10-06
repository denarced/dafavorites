package deviantart

import (
	"fmt"
	"strings"
	"testing"

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
