package xkcdslack

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"appengine"
	"appengine/datastore"
	"appengine/search"
)

type searchSlashCommandAttachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
	Text      string `json:"text"`
	ImageUrl  string `json:"image_url"`
}

type searchSlashCommandResponse struct {
	Text         string                          `json:"text,omitempty"`
	ResponseType string                          `json:"response_type,omitempty"`
	Attachments  []*searchSlashCommandAttachment `json:"attachments,omitempty"`
}

func searchSlashCommandHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var comicList []*ComicSearch
	for t := index.Search(c, text, nil); ; {
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

	sr := new(searchSlashCommandResponse)
	if len(comicList) == 0 {
		sr.Text = "EPOCH FAIL!"
	} else {
		n := rand.Intn(len(comicList))
		comic := comicList[n]

		attachment := &searchSlashCommandAttachment{
			Text:      comic.Alt,
			TitleLink: fmt.Sprintf("https://xkcd.com/%s/", comic.Num),
			Title:     comic.Title,
			ImageUrl:  comic.Img,
		}
		sr.Attachments = []*searchSlashCommandAttachment{attachment}
		sr.ResponseType = "in_channel"
	}
	rsp, _ := json.Marshal(sr)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, string(rsp))

	logSearch(c, r.FormValue("user_name"), r.FormValue("channel_name"), r.FormValue("team_domain"), text)
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
