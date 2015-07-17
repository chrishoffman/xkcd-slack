package xkcdslack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"appengine"
	"appengine/search"
	"appengine/urlfetch"
)

type searchWebhookResponse struct {
	Text string `json:"text"`
}

func searchWebhookHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	triggerWord := r.FormValue("trigger_word")
	text := r.FormValue("text")
	if triggerWord == "" || text == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	searchText := strings.Replace(text, triggerWord, "", 1)
	comic := searchIndex(c, w, searchText)
	if comic == nil {
		return
	}

	xkcdURL := fmt.Sprintf("https://xkcd.com/%s/", comic.Num)
	sr := &searchWebhookResponse{xkcdURL}
	rsp, _ := json.Marshal(sr)
	fmt.Fprintf(w, string(rsp))
}

type searchSlashResponse struct {
	Channel     string `json:"channel"`
	Text        string `json:"text"`
	UnfurlLinks bool   `json:"unfurl_links"`
}

func searchSlashHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	searchText := r.FormValue("text")
	if searchText == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	comic := searchIndex(c, w, searchText)
	if comic == nil {
		return
	}

	xkcdURL := fmt.Sprintf("https://xkcd.com/%s/", comic.Num)
	if callback := r.FormValue("callback"); callback == "" {
		fmt.Fprintf(w, xkcdURL)
	} else {
		sr := &searchSlashResponse{
			Text:        fmt.Sprintf("<%s>", xkcdURL),
			UnfurlLinks: true,
			Channel:     r.FormValue("channel_name"),
		}
		rsp, _ := json.Marshal(sr)

		client := urlfetch.Client(c)
		req, _ := http.NewRequest("POST", callback, bytes.NewBuffer(rsp))
		_, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error posting to callback", http.StatusInternalServerError)
		}
	}
}

func searchIndex(c appengine.Context, w http.ResponseWriter, searchText string) *ComicSearch {
	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	var comicList []*ComicSearch
	for t := index.Search(c, searchText, nil); ; {
		var xkcd ComicSearch
		_, err := t.Next(&xkcd)
		if err == search.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}

		comicList = append(comicList, &xkcd)
	}

	if len(comicList) > 0 {
		n := rand.Intn(len(comicList))
		return comicList[n]
	}

	http.Error(w, "No match", http.StatusNotFound)
	return nil
}
