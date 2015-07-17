package xkcdslack

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func init() {
	router := httprouter.New()
	router.POST("/search/webhook", searchWebhookHandler)
	router.POST("/search/slash", searchSlashHandler)
	router.POST("/index", index)
	router.GET("/task/backfill", backfill)
	http.Handle("/", router)
}
