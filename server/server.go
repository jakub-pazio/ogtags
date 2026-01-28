package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jakub-pazio/ogtags/cache"
	"github.com/jakub-pazio/ogtags/httpclient"
	"github.com/jakub-pazio/ogtags/tags"
)

type Server struct {
	c      cache.Cache
	httpc  httpclient.HttpClient
	Mux    *http.ServeMux
	logger *slog.Logger
}

func New(c cache.Cache, httpc httpclient.HttpClient, indexFile []byte, logger *slog.Logger) *Server {
	s := &Server{
		c:      c,
		httpc:  httpc,
		Mux:    http.NewServeMux(),
		logger: logger,
	}

	indexHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s.logger.InfoContext(ctx, "Client requested index page")
		w.Write(indexFile)
	})

	s.Mux.Handle("/", indexHandler)
	s.Mux.Handle("/tags", s.tagsHandlerProvider())

	return s
}

func (s *Server) tagsHandlerProvider() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx        = r.Context()
			urlFromReq = r.URL.Query().Get("url")
		)

		if urlFromReq == "" {
			s.logger.WarnContext(ctx, "No url provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ok, cachedTags, err := s.c.GetRequiredTags(ctx, urlFromReq)
		if err != nil {
			s.logger.WarnContext(ctx, "Could not get tags from cache",
				"error", err,
				"url", urlFromReq,
			)
		}

		if ok {
			w.Write([]byte(cachedTags))
			return
		}

		body, err := s.httpc.Fetch(ctx, urlFromReq)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to fetch body",
				"error", err,
				"url", urlFromReq,
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		parsedTags, err := tags.ParseTags(body)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed parsing url",
				"error", err,
				"url", urlFromReq,
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		requiredTags := tags.GetRequiredTags(parsedTags)
		json.NewEncoder(w).Encode(requiredTags)

		bs, err := json.Marshal(requiredTags)
		if err != nil {
			s.logger.ErrorContext(ctx, "Could not marshal tags",
				"error", err,
				"tags", requiredTags,
			)
			return
		}

		s.logger.InfoContext(ctx, "Send og tags",
			"url", urlFromReq,
		)

		s.c.SetRequiredTags(ctx, urlFromReq, string(bs))
	})
}
