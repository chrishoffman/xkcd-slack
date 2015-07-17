package xkcdslack

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"appengine"
	"appengine/search"
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
	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	if len(comicList) == 0 {
		http.Error(w, "No match", http.StatusNotFound)
		return
	}

	n := rand.Intn(len(comicList))
	xkcdURL := fmt.Sprintf("https://xkcd.com/%s/", comicList[n].Num)
	sr := &searchWebhookResponse{xkcdURL}
	rsp, _ := json.Marshal(sr)
	fmt.Fprintf(w, string(rsp))
}
