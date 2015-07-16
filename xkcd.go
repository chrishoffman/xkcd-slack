package xkcdslack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"appengine"
	"appengine/urlfetch"
)

// Comic is a struct that contains infomation about an xkcd comic.
type Comic struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	SafeTitle  string `json:"safe_title"`
	Img        string `json:"img"`
	Alt        string `json:"alt"`
	Year       string `json:"year"`
	Month      string `json:"month"`
	Day        string `json:"day"`
	News       string `json:"news"`
	Link       string `json:"link"`
	Transcript string `json:"transcript"`
}

// New reads from an io.Reader and returns a *Comic struct.
func New(r io.Reader) (*Comic, error) {
	d := json.NewDecoder(r)
	c := new(Comic)
	err := d.Decode(c)
	return c, err
}

const (
	currentURL  = "https://xkcd.com/info.0.json"
	templateURL = "https://xkcd.com/%v/info.0.json"
)

// Get fetches information about the xkcd comic number `n'.
func Get(c appengine.Context, n int) (*Comic, error) {
	url := fmt.Sprintf(templateURL, n)
	return getByURL(c, url)
}

// GetCurrent fetches information for the newest xkcd comic.
func GetCurrent(c appengine.Context) (*Comic, error) {
	return getByURL(c, currentURL)
}

// getByURL returns infomation downloaded from `url'.
func getByURL(c appengine.Context, url string) (*Comic, error) {
	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New(resp.Status)
	}
	return New(resp.Body)
}
