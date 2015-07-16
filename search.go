package xkcdslack

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"appengine"
	"appengine/search"
)

type SearchResponse struct {
	Text string `json:"text"`
}

func init() {
	http.HandleFunc("/search", searchHandler)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "GET/POST requests only", http.StatusMethodNotAllowed)
		return
	}

	c := appengine.NewContext(r)

	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	triggerWord := r.FormValue("trigger_word")
	text := r.FormValue("text")
	if triggerWord == "" || text == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var comicList []*ComicSearch
	searchText := strings.Replace(text, triggerWord, "", 1)
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
		xkcdURL := fmt.Sprintf("https://xkcd.com/%s/", comicList[n].Num)
		sr := &SearchResponse{xkcdURL}
		rsp, _ := json.Marshal(sr)
		fmt.Fprintf(w, string(rsp))
		return
	}

	http.Error(w, "No match", http.StatusNotFound)
}
