package dafavorites

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	djson "github.com/denarced/dafavorites/lib/dafavorites/json"
	"github.com/denarced/dafavorites/shared/shared"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToRssFile(t *testing.T) {
	// SETUP SUT
	shared.InitTestLogging(t)
	rssBytes, err := os.ReadFile("testdata/rss.xml")
	req := assert.New(t)
	req.Nil(err)

	// EXERCISE
	rssFile, err := toRssFile(rssBytes)

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
			shared.InitTestLogging(t)
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

func TestFetchFavorites(t *testing.T) {
	shared.InitTestLogging(t)
	dirp := "/root"
	httpClient := newTestHTTPClient()
	fsys := &afero.Afero{Fs: afero.NewMemMapFs()}
	ctx := &TestContext{
		fsys:       fsys,
		username:   "denarced",
		httpClient: httpClient,
	}

	// EXERCISE
	fetched := FetchFavorites(dirp, 1, ctx)

	// VERIFY
	ass := assert.New(t)
	deviations := fetched.SavedDeviations
	annaURL := "https://www.deviantart.com/davidcraigellis/art/Anna-Rose-13-1079160547"
	ass.Equal(
		[]djson.RssItem{
			{
				Title:           "Anna Rose 13",
				Link:            annaURL,
				GUID:            annaURL,
				PublicationDate: "Thu, 25 Jul 2024 19:44:49 PDT",
				Author:          "DavidCraigEllis",
				URL:             "https://images-wixmp.wixmp.com/anna.jpg",
				Dimensions: djson.Dimensions{
					Width:  894,
					Height: 894,
				},
			},
			{
				Title:           "Kat",
				Link:            "https://www.deviantart.com/friesellfly/art/Kat-1042398875",
				GUID:            "https://www.deviantart.com/friesellfly/art/Kat-1042398875",
				PublicationDate: "Mon, 15 Apr 2024 08:29:36 PDT",
				Author:          "FriesellFly",
				URL:             "https://images-wixmp.com/kat.jpg",
				Dimensions: djson.Dimensions{
					Width:  730,
					Height: 1095,
				},
			},
		},
		[]djson.RssItem{deviations[0].RssItem, deviations[1].RssItem},
	)
	verifyFileContent(require.New(t), fsys, dirp, "anna.jpg", []byte("anna\n"))
	verifyFileContent(require.New(t), fsys, dirp, "kat.jpg", []byte("kat\n"))
	ass.NotNil(fetched.Timestamp)
	ass.Nil(httpClient.err)
}

type TestContext struct {
	fsys       *afero.Afero
	httpClient *TestHTTPClient
	username   string
}

func (v *TestContext) Fsys() *afero.Afero {
	return v.fsys
}

func (v *TestContext) Username() string {
	return v.username
}

func (v *TestContext) CreateClient() HTTPClient {
	return v.httpClient
}

type TestHTTPClient struct {
	err error
}

func newTestHTTPClient() *TestHTTPClient {
	return &TestHTTPClient{}
}

func (v *TestHTTPClient) Fetch(url string) ([]byte, error) {
	filep := filepath.Join("testdata", "TestFetchFavorites", strings.ReplaceAll(url, "/", "_"))
	bytes, err := readFile(filep)
	if v.err == nil && err != nil {
		v.err = err
	}
	return bytes, err
}

func readFile(filep string) ([]byte, error) {
	file, err := os.Open(filep)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	var byteCount int64
	{
		info, err := os.Stat(filep)
		if err != nil {
			return []byte{}, err
		}
		byteCount = info.Size()
	}

	bytes := make([]byte, byteCount)
	n, err := file.Read(bytes)
	if err != nil {
		return []byte{}, err
	}
	if int64(n) != byteCount {
		return []byte{}, fmt.Errorf(
			"unexpected number of bytes read: %d vs %d, file: %s",
			byteCount,
			n,
			filep,
		)
	}
	return bytes, nil
}

func verifyFileContent(
	req *require.Assertions,
	fsys *afero.Afero,
	rootDir, filename string,
	expected []byte,
) {
	var foundPath string
	err := fsys.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		req.Nil(err)
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == filename {
			foundPath = path
		}
		return nil
	})
	req.Nil(err, "fsys-walk-err-return")
	req.NotEmpty(foundPath, "filepath-not-found")
	bytes, err := fsys.ReadFile(foundPath)
	req.Nil(err, "read-file-err")
	req.Equal(expected, bytes)
}
