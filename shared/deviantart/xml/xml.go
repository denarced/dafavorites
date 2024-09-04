// Package xml contains types involved in XML marshaling.
// Dedicated package exists merely to make it easier to track which types have to have capitalized
// names.
// revive:disable:max-public-structs Pointless to have a limit. Go forces everything in this package
// to be public because otherwise XML unmarshaling doesn't work.
// revive:disable:line-length-limit The XML examples are useful and there's no point in breaking
// them down to multiple lines.
package xml

import "encoding/xml"

// RssElement is the root element of Deviant Art's RSS xml
type RssElement struct {
	XMLName xml.Name `xml:"rss"`
	// The single channel element in the xml. At least no more than one hasn't
	// been seen during development.
	Channel ChannelElement `xml:"channel"`
}

// ChannelElement is the single channel element in Deviant Art RSS xml
// that's located inside the root rss element.
type ChannelElement struct {
	// The link elements. In this bunch we're mostly interested in the "next"
	// links.
	Links []LinkElement `xml:"link"`
	// The actual item elements, each of which contains a single favorite
	// deviation.
	RssItems []RssItemElement `xml:"item"`
}

// LinkElement is a URL to another Deviant Art RSS xml
type LinkElement struct {
	// Relation, e.g. "next". Each RSS xml contains x amount of favorite items
	// and then the URL in "next" contains the next RSS xml that contains more.
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// RssItemElement is a single <item> in deviant art RSS xml.
// Contains information on any given favorite deviation.
//
// Example:
//
//	   <item>
//		   <title>MODEL NO. TH-X11-38</title>
//		   <link>http://wojtekfus.deviantart.com/art/MODEL-NO-TH-X11-38-577869970</link>
//		   <guid isPermaLink="true">http://wojtekfus.deviantart.com/art/MODEL-NO-TH-X11-38-577869970</guid>
//		   <pubDate>Sun, 13 Dec 2015 16:14:02 PST</pubDate>
//		   <media:title type="plain">MODEL NO. TH-X11-38</media:title>
//		   <media:keywords></media:keywords>
//		   <media:rating>nonadult</media:rating>
//		   <media:category label="Sci-Fi">digitalart/paintings/scifi</media:category>
//		   <media:credit role="author" scheme="urn:ebu">WojtekFus</media:credit>
//		   <media:credit role="author" scheme="urn:ebu">http://a.deviantart.net/avatars/w/o/wojtekfus.jpg?5</media:credit>
//		   <media:copyright url="http://wojtekfus.deviantart.com">Copyright 2015 WojtekFus</media:copyright>
//		   <media:description type="html">Image done for the workshop in Taipei that&#039;s going to happen next week:&amp;nbsp;&lt;a class=&quot;external&quot; href=&quot;http://www.deviantart.com/users/outgoing?http://www.likmeetup.com/&quot;&gt;www.likmeetup.com/&lt;/a&gt;&lt;br /&gt;&lt;br /&gt;Struggled with it a lot myself, but I&#039;m letting it go, haha &lt;img src=&quot;http://e.deviantart.net/emoticons/s/smile.gif&quot; width=&quot;15&quot; height=&quot;15&quot; alt=&quot;:)&quot; data-embed-type=&quot;emoticon&quot; data-embed-id=&quot;391&quot; title=&quot;:) (Smile)&quot;/&gt; Hope you like it! Love you guys!</media:description>            <media:thumbnail url="http://t02.deviantart.net/FeGyfpR_tb8vGcOarQm_dyvBd7U=/fit-in/150x150/filters:no_upscale():origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="84" width="150"/>            <media:thumbnail url="http://t14.deviantart.net/BPI5k4O3FXvK1PF7VJXREIujS0I=/fit-in/300x900/filters:no_upscale():origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="169" width="300"/>            <media:thumbnail url="http://t15.deviantart.net/ZS17sMYJv1Whk_q1lyP4DrvdH30=/300x200/filters:fixed_height(100,100):origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="169" width="300"/>
//		   <media:content url="http://pre03.deviantart.net/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg" height="670" width="1192" medium="image"/>
//		   <description>Image done for the workshop in Taipei that&#039;s going to happen next week:&amp;nbsp;&lt;a class=&quot;external&quot; href=&quot;http://www.deviantart.com/users/outgoing?http://www.likmeetup.com/&quot;&gt;www.likmeetup.com/&lt;/a&gt;&lt;br /&gt;&lt;br /&gt;Struggled with it a lot myself, but I&#039;m letting it go, haha &lt;img src=&quot;http://e.deviantart.net/emoticons/s/smile.gif&quot; width=&quot;15&quot; height=&quot;15&quot; alt=&quot;:)&quot; data-embed-type=&quot;emoticon&quot; data-embed-id=&quot;391&quot; title=&quot;:) (Smile)&quot;/&gt; Hope you like it! Love you guys!&lt;br /&gt;&lt;div&gt;&lt;img src=&quot;http://t15.deviantart.net/ZS17sMYJv1Whk_q1lyP4DrvdH30=/300x200/filters:fixed_height(100,100):origin()/pre03/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg&quot; alt=&quot;thumbnail&quot; /&gt;&lt;/div&gt;</description>
//	   </item>
type RssItemElement struct {
	Title           string              `xml:"title"`
	Link            string              `xml:"link"`
	GUID            string              `xml:"guid"`
	PublicationDate string              `xml:"pubDate"`
	URL             string              `xml:"url"`
	Width           int                 `xml:"width"`
	Height          int                 `xml:"height"`
	Credits         []ItemCreditElement `xml:"credit"`
	Content         ItemContentElement  `xml:"content"`
}

// ItemContentElement in deviant art RSS xml.
// Example:
//
//	<media:content
//	    url="http://pre03.deviantart.net/bbec/th/pre/f/2015/347/b/f/model_no__th_x11_38_by_wojtekfus-d9k1rbm.jpg"
//	    height="670"
//	    width="1192"
//	    medium="image"/>
type ItemContentElement struct {
	URL    string `xml:"url,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

// ItemCreditElement is a credit element in Deviant Art RSS xml.
// Example:
//
//	<media:credit role="author" scheme="urn:ebu">WojtekFus</media:credit>
type ItemCreditElement struct {
	Role  string `xml:"role,attr"`
	Value string `xml:",chardata"`
}
