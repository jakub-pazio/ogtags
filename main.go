package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jakub-pazio/ogtags/cache"
	"github.com/jakub-pazio/ogtags/httpclient"
	"github.com/jakub-pazio/ogtags/server"
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

	tagsCache, cacheShutdownFn := cache.New()
	defer cacheShutdownFn()

	indexFile, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		log.Panicf("Error reading index.html: %v\n", err)
	}

	hc := httpclient.New()

	s := server.New(tagsCache, hc, indexFile)

	fmt.Printf("Starting application on port %s\n", *portFlag)

	appPort := ":" + *portFlag
	if err := http.ListenAndServe(appPort,
		otelhttp.NewHandler(&s.Mux, "ogtags-server"),
	); err != nil {
		log.Println("Server failed", err)
	}
}
