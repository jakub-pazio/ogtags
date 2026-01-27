package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jakub-pazio/ogtags/tags"
	"github.com/jakub-pazio/ogtags/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed static/index.html
var staticFiles embed.FS

var portFlag = flag.String("p", "8081", "port to run application at")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	ctx := context.Background()
	shutdown, err := telemetry.SetupOtelSDK(ctx)
	if err != nil {
		fmt.Println("Could not setup telemetry", err)
		os.Exit(1)
	}
	defer shutdown(ctx)

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
		ts, err := tags.GetOgMetaTags(r.Context(), reqUrl)
		if err != nil {
			log.Printf("Error getting OG tags: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rts := tags.GetRequiredTags(r.Context(), ts)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rts)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexFile)
	})

	handler := otelhttp.NewHandler(
		mux,
		"/",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)

	appPort := ":" + *portFlag
	if err := http.ListenAndServe(appPort, handler); err != nil {
		log.Println("Server failed", err)
	}
}
