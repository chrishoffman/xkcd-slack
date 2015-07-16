package xkcdslack

import (
	"fmt"
	"net/http"
	"strconv"

	"appengine"
	"appengine/search"
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
}

func index(w http.ResponseWriter, r *http.Request) {
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
