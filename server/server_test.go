package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jakub-pazio/ogtags/cache"
	"github.com/jakub-pazio/ogtags/httpclient"
)

func TestIntegrationTagsHandler(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testCache := &fakeCache{}
	testClient := &fakeHttpClient{}
	testIndexFile := []byte("response")
	server := New(testCache, testClient, testIndexFile, testLogger)

	ts := httptest.NewServer(server.Mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/tags?url=https://fakepage.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("error reading body: %v", err)
	}

	want := "{\"title\":{\"name\":\"og:title\",\"value\":\"The Rock\"},\"type\":{\"name\":\"og:type\",\"value\":\"video.movie\"},\"image\":{\"name\":\"og:image\",\"value\":\"https://ia.media-imdb.com/images/rock.jpg\"},\"url\":{\"name\":\"og:url\",\"value\":\"https://www.imdb.com/title/tt0117500/\"}}\n"
	got := string(body)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("body mismatch (-want +got):\n%s", diff)
	}
}

var _ cache.Cache = (*fakeCache)(nil)

type fakeCache struct {
}

func (f *fakeCache) GetRequiredTags(ctx context.Context, url string) (bool, string, error) {
	return false, "", nil
}

func (f *fakeCache) SetRequiredTags(ctx context.Context, url string, body string) error {
	return nil
}

var _ httpclient.HttpClient = (*fakeHttpClient)(nil)

type fakeHttpClient struct {
}

const fbody = `
<html prefix="og: https://ogp.me/ns#">
<head>
<title>The Rock (1996)</title>
<meta property="og:title" content="The Rock" />
<meta property="og:type" content="video.movie" />
<meta property="og:url" content="https://www.imdb.com/title/tt0117500/" />
<meta property="og:image" content="https://ia.media-imdb.com/images/rock.jpg" />
</head>
<body>
	<h1>Some page</h1>
</body>
</html>
`

func (f *fakeHttpClient) Fetch(ctx context.Context, url string) (string, error) {
	return fbody, nil
}
