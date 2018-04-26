// Package coverart fetches coverart for albums
package coverart

import (
	"net/url"
)

// Fetcher fetches cover art
type Fetcher interface {
	Fetch(id []string) (map[string]Image, error)
}

// Image is an album image
type Image struct {
	ID          string
	URL         *url.URL
	Description description
	Height      int // in pixels
	Width       int // in pixels
}

type description string

// Front is the main album image
const Front description = "front"

// Back is the back image of an album
const Back description = "back"
