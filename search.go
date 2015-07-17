package xkcdslack

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"appengine"
	"appengine/datastore"
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

	sr := new(searchWebhookResponse)
	if len(comicList) == 0 {
		sr.Text = "EPOCH FAIL!"
	} else {
		n := rand.Intn(len(comicList))
		sr.Text = fmt.Sprintf("https://xkcd.com/%s/", comicList[n].Num)
	}
	rsp, _ := json.Marshal(sr)
	fmt.Fprintf(w, string(rsp))

	logSearch(c, r.FormValue("user_name"), r.FormValue("channel_name"), r.FormValue("team_domain"), searchText)
}

type searchLog struct {
	Date       time.Time
	User       string
	Channel    string
	Domain     string
	SearchTerm string
}

func logSearch(c appengine.Context, user, channel, domain, searchTerm string) {
	log := searchLog{
		User:       user,
		Channel:    channel,
		Domain:     domain,
		SearchTerm: searchTerm,
		Date:       time.Now(),
	}
	datastore.Put(c, datastore.NewIncompleteKey(c, "SearchLog", nil), &log)
}
