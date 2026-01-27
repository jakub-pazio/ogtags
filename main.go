package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jakub-pazio/ogtags/tags"
)

func main() {
	mux := http.NewServeMux()

	indexFile, err := os.ReadFile("./static/index.html")
	if err != nil {
		fmt.Println("Error reading index.html", err)
		os.Exit(1)
	}

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		reqUrl := r.URL.Query().Get("url")
		if reqUrl == "" {
			fmt.Println("No url in req")
			w.Write([]byte("Add url\n"))
			return
		}

		//TODO: Check if input is correct
		ts, err := tags.GetOgMetaTags(reqUrl)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("Error", err)
			return
		}

		rts := tags.GetRequiredTags(ts)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rts)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexFile)
	})

	if err := http.ListenAndServe(":8081", mux); err != nil {
		fmt.Println("Server failed", err)
	}
}
