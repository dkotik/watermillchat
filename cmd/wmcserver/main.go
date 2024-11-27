package main

import (
	"net/http"
	"watermillchat/datastar"
)

func main() {
	http.HandleFunc("/datastar.js", datastar.SourceHandler)
	http.HandleFunc("/datastar.js.map", datastar.SourceMapHandler)
	http.HandleFunc("/events.html", EventHandler)
	http.HandleFunc("/index.html", IndexHandler)
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe("localhost:8081", nil)
}
