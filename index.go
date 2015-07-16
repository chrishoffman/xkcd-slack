package xkcdslack

import (
	"fmt"
	"net/http"
	"strconv"

	"appengine"
	"appengine/search"
	"appengine/taskqueue"
)

const xkcdIndex = "xkcd"

type ComicSearch struct {
	Num   string
	Title string
	Img   string
	Alt   string
}

func init() {
	http.HandleFunc("/index", index)
	http.HandleFunc("/backfill", backfill)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST requests only", http.StatusMethodNotAllowed)
		return
	}

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
	}

	id := strconv.Itoa(x.Num)
	_, err = index.Put(c, id, xSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, "Retrieved document: ", xSearch)
}

func backfill(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST requests only", http.StatusMethodNotAllowed)
		return
	}

	c := appengine.NewContext(r)

	index, err := search.Open(xkcdIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	current, _ := GetCurrent(c)
	for i := 1; i < current.Num; i++ {
		var s ComicSearch
		err := index.Get(c, strconv.Itoa(i), &s)
		if err == nil {
			continue
		}

		t := taskqueue.NewPOSTTask("/index", map[string][]string{"id": {strconv.Itoa(i)}})
		if _, err := taskqueue.Add(c, t, ""); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
