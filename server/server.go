package server

import (
	"encoding/json"
	"net/http"

	"github.com/jakub-pazio/ogtags/cache"
	"github.com/jakub-pazio/ogtags/httpclient"
	"github.com/jakub-pazio/ogtags/tags"
)

type Server struct {
	c     cache.Cache
	httpc httpclient.HttpClient
	Mux   http.ServeMux
}

func New(c cache.Cache, httpc httpclient.HttpClient, indexFile []byte) *Server {
	s := &Server{
		c:     c,
		httpc: httpc,
		Mux:   *http.NewServeMux(),
	}

	indexHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexFile)
	})

	s.Mux.Handle("/", indexHandler)
	s.Mux.Handle("/tags", s.tagsHandler())

	return s
}

func (s *Server) tagsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx        = r.Context()
			urlFromReq = r.URL.Query().Get("url")
		)

		if urlFromReq == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ok, cachedTags, err := s.c.GetRequiredTags(ctx, urlFromReq)
		if err != nil {
			//TODO: log errors
		}

		if ok {
			w.Write([]byte(cachedTags))
			return
		}

		body, err := s.httpc.Fetch(ctx, urlFromReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		parsedTags, err := tags.ParseTags(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		requiredTags := tags.GetRequiredTags(parsedTags)
		json.NewEncoder(w).Encode(requiredTags)

		bs, err := json.Marshal(requiredTags)
		if err != nil {
			//TODO: logging
			return
		}

		s.c.SetRequiredTags(ctx, urlFromReq, string(bs))
	})
}
