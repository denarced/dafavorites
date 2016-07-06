package deviantart

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const logFlags = log.LstdFlags | log.Lshortfile

var (
	infoLogger  = log.New(os.Stdout, "INFO ", logFlags)
	errorLogger = log.New(os.Stderr, "ERROR ", logFlags)
)

// Dimensions of the deviation
type Dimensions struct {
	Width  int
	Height int
}

// Deviation item
type RssItem struct {
	// I.e. the name of the deviation
	Title string
	// URL to the deviation, usually identical to Guid
	Link            string
	Guid            string
	PublicationDate string
	Author          string
	Url             string
	Dimensions      Dimensions
}

// Items of the one Deviant Art RSS file and the next one's URL
type RssFile struct {
	NextUrl  string
	RssItems []RssItem
}

// URL to another Deviant Art RSS xml
type LinkElement struct {
	// Relation, e.g. "next". Each RSS xml contains x amount of favorite items
	// and then the URL in "next" contains the next RSS xml that contains more.
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// A credit element in Deviant Art RSS xml.
// Example: <media:credit role="author" scheme="urn:ebu">WojtekFus</media:credit>
type ItemCreditElement struct {
	Role  string `xml:"role,attr"`
	Value string `xml:",chardata"`
}

// Example:
//     <media:content
//         url="http://pre03.deviantart.net/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg"
//         height="670"
//         width="1192"
//         medium="image"/>
type ItemContentElement struct {
	Url    string `xml:"url,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

// Contains information on any given favorite deviation.
//
// Example:
//     <item>
//  	   <title>MODEL NO. TH-X11-38</title>
//  	   <link>http://wojtekfus.deviantart.com/art/MODEL-NO-TH-X11-38-577869970</link>
//  	   <guid isPermaLink="true">http://wojtekfus.deviantart.com/art/MODEL-NO-TH-X11-38-577869970</guid>
//  	   <pubDate>Sun, 13 Dec 2015 16:14:02 PST</pubDate>
//  	   <media:title type="plain">MODEL NO. TH-X11-38</media:title>
//  	   <media:keywords></media:keywords>
//  	   <media:rating>nonadult</media:rating>
//  	   <media:category label="Sci-Fi">digitalart/paintings/scifi</media:category>
//  	   <media:credit role="author" scheme="urn:ebu">WojtekFus</media:credit>
//  	   <media:credit role="author" scheme="urn:ebu">http://a.deviantart.net/avatars/w/o/wojtekfus.jpg?5</media:credit>
//  	   <media:copyright url="http://wojtekfus.deviantart.com">Copyright 2015 WojtekFus</media:copyright>
//  	   <media:description type="html">Image done for the workshop in Taipei that&#039;s going to happen next week:&amp;nbsp;&lt;a class=&quot;external&quot; href=&quot;http://www.deviantart.com/users/outgoing?http://www.likmeetup.com/&quot;&gt;www.likmeetup.com/&lt;/a&gt;&lt;br /&gt;&lt;br /&gt;Struggled with it a lot myself, but I&#039;m letting it go, haha &lt;img src=&quot;http://e.deviantart.net/emoticons/s/smile.gif&quot; width=&quot;15&quot; height=&quot;15&quot; alt=&quot;:)&quot; data-embed-type=&quot;emoticon&quot; data-embed-id=&quot;391&quot; title=&quot;:) (Smile)&quot;/&gt; Hope you like it! Love you guys!</media:description>            <media:thumbnail url="http://t02.deviantart.net/FeGyfpR_tb8vGcOarQm_dyvBd7U=/fit-in/150x150/filters:no_upscale():origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="84" width="150"/>            <media:thumbnail url="http://t14.deviantart.net/BPI5k4O3FXvK1PF7VJXREIujS0I=/fit-in/300x900/filters:no_upscale():origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="169" width="300"/>            <media:thumbnail url="http://t15.deviantart.net/ZS17sMYJv1Whk_q1lyP4DrvdH30=/300x200/filters:fixed_height(100,100):origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="169" width="300"/>
//  	   <media:content url="http://pre03.deviantart.net/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="670" width="1192" medium="image"/>
//  	   <description>Image done for the workshop in Taipei that&#039;s going to happen next week:&amp;nbsp;&lt;a class=&quot;external&quot; href=&quot;http://www.deviantart.com/users/outgoing?http://www.likmeetup.com/&quot;&gt;www.likmeetup.com/&lt;/a&gt;&lt;br /&gt;&lt;br /&gt;Struggled with it a lot myself, but I&#039;m letting it go, haha &lt;img src=&quot;http://e.deviantart.net/emoticons/s/smile.gif&quot; width=&quot;15&quot; height=&quot;15&quot; alt=&quot;:)&quot; data-embed-type=&quot;emoticon&quot; data-embed-id=&quot;391&quot; title=&quot;:) (Smile)&quot;/&gt; Hope you like it! Love you guys!&lt;br /&gt;&lt;div&gt;&lt;img src=&quot;http://t15.deviantart.net/ZS17sMYJv1Whk_q1lyP4DrvdH30=/300x200/filters:fixed_height(100,100):origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg&quot; alt=&quot;thumbnail&quot; /&gt;&lt;/div&gt;</description>
//     </item>
type RssItemElement struct {
	Title           string              `xml:"title"`
	Link            string              `xml:"link"`
	Guid            string              `xml:"guid"`
	PublicationDate string              `xml:"pubDate"`
	Url             string              `xml:"url"`
	Width           int                 `xml:"width"`
	Height          int                 `xml:"height"`
	Credits         []ItemCreditElement `xml:"credit"`
	Content         ItemContentElement  `xml:"content"`
}

// The single channel element in Deviant Art RSS xml that's located inside the
// root rss element.
type ChannelElement struct {
	// The link elements. In this bunch we're mostly interested in the "next"
	// links.
	Links []LinkElement `xml:"link"`
	// The actual item elements, each of which contains a single favorite
	// deviation.
	RssItems []RssItemElement `xml:"item"`
}

// The root element of Deviant Art's RSS xml
type RssElement struct {
	XMLName xml.Name `xml:"rss"`
	// The single channel element in the xml. At least no more than one hasn't
	// been seen during development.
	Channel ChannelElement `xml:"channel"`
}

// Convert deviant art structures to our own
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
			Title:           each.Title,
			Link:            each.Link,
			Guid:            each.Guid,
			PublicationDate: each.PublicationDate,
			Author:          author,
			Url:             each.Content.Url,
			Dimensions: Dimensions{
				Width:  each.Content.Width,
				Height: each.Content.Height,
			},
		})
	}
	return rssItems
}

// Fetch deviant art RSS from url
func FetchRssFile(url string) (RssFile, error) {
	infoLogger.Println("Fetch RSS file:", url)
	resp, err := http.Get(url)
	if err != nil {
		errorLogger.Println("Failed to fetch RSS file:", err)
		return RssFile{}, err
	}
	defer resp.Body.Close()
	contentBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorLogger.Println("Failed to read fetched rss file:", err)
		return RssFile{}, err
	}

	rssElement := RssElement{}
	xml.Unmarshal(contentBytes, &rssElement)
	rssItems := itemElementsToItems(rssElement.Channel.RssItems)

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
	}, nil
}
