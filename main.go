package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/jakub-pazio/ogtags/tags"
)

//go:embed static/index.html
var staticFiles embed.FS

var portFlag = flag.String("p", "8081", "port to run application at")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	indexFile, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		log.Panicf("Error reading index.html: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		reqUrl := r.URL.Query().Get("url")
		if reqUrl == "" {
			log.Println("No url in request")
			w.Write([]byte("Add url\n"))
			return
		}

		//TODO: Check if input is correct
		ts, err := tags.GetOgMetaTags(reqUrl)
		if err != nil {
			log.Printf("Error getting OG tags: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rts := tags.GetRequiredTags(ts)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rts)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexFile)
	})

	appPort := ":" + *portFlag
	if err := http.ListenAndServe(appPort, mux); err != nil {
		log.Println("Server failed", err)
	}
}
