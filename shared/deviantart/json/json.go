// Package json contains types involved in JSON marshaling.
// Dedicated package exists merely to make it easier to track which types have to have capitalized
// names.
// revive:disable:max-public-structs Pointless to have a limit. Go forces everything in this package
// to be public because otherwise JSON marshaling doesn't work.
package json

import "time"

// DeviantFetch is one full fetch, all deviations, their saved filenames etc.
type DeviantFetch struct {
	SavedDeviations []SavedDeviation
	Timestamp       time.Time
}

// SavedDeviation is a single saved deviation
type SavedDeviation struct {
	RssItem  RssItem
	Filename string
}

// RssItem is a single <item> in deviant art RSS
type RssItem struct {
	// I.e. the name of the deviation
	Title string
	// URL to the deviation, usually identical to GUID
	Link            string
	GUID            string
	PublicationDate string
	Author          string
	URL             string
	Dimensions      Dimensions
}

// Dimensions of the deviation
type Dimensions struct {
	Width  int
	Height int
}
