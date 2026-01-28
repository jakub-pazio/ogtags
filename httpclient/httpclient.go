package httpclient

import (
	"context"
	"io"
	"net/http"
)

type HttpClient interface {
	Fetch(ctx context.Context, url string) (string, error)
}

type client struct {
}

// Fetch implements HttpClient.
func (c *client) Fetch(ctx context.Context, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

var _ HttpClient = (*client)(nil)

func New() HttpClient {
	return &client{}
}
