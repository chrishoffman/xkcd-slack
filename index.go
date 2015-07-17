package xkcdslack

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"appengine"
	"appengine/search"
	"appengine/taskqueue"
)

const xkcdIndex = "xkcd"

type ComicSearch struct {
	Num        string
	Title      string
	Img        string
	Alt        string
	Transcript string
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var x *Comic
	comicNum := r.FormValue("id")
	if comicNum != "" {
		iComicNum, _ := strconv.Atoi(comicNum)
		x, _ = Get(c, iComicNum)
	} else {
		x, _ = GetCurrent(c)
	}
	xSearch := &ComicSearch{
		strconv.Itoa(x.Num),
		x.Title,
		x.Img,
		x.Alt,
		x.Transcript,
	}

	id := strconv.Itoa(x.Num)
	_, err = index.Put(c, id, xSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func backfill(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	current, _ := GetCurrent(c)
	for i := 1; i <= current.Num; i++ {
		// xcdc returns 404 with issue 404
		if i == 404 {
			continue
		}

		comicNum := strconv.Itoa(i)
		var s ComicSearch
		err := index.Get(c, comicNum, &s)
		if err == nil {
			continue
		}

		t := taskqueue.NewPOSTTask("/index", map[string][]string{"id": {comicNum}})
		if _, err := taskqueue.Add(c, t, ""); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
