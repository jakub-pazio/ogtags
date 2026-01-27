package tags

import (
	"context"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	OgTitle string = "og:title"
	OgType  string = "og:type"
	OgImage string = "og:image"
	OgUrl   string = "og:url"
)

const name = "ogtags"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RequiredTags struct {
	Title Tag `json:"title"`
	Type  Tag `json:"type"`
	Image Tag `json:"image"`
	Url   Tag `json:"url"`
}

func Parse(s string) Tag {
	if s == "" {
		return Tag{}
	}

	return Tag{
		Name:  "Title",
		Value: "Post about OCaml",
	}
}

func GetBody(ctx context.Context, url string) (string, error) {
	ctx, span := tracer.Start(ctx, "get-body")
	defer span.End()

	logger.InfoContext(ctx, "getting body page", "url", url)

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

func getOgTags(body string) ([]Tag, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var res []Tag

	for n := range doc.Descendants() {
		if isNodeOgTag(n) {
			t := getOgTagValues(n)
			res = append(res, t)
		}
	}

	return res, nil
}

func isNodeOgTag(n *html.Node) bool {
	if n.Type != html.ElementNode || n.DataAtom != atom.Meta {
		return false
	}
	for _, a := range n.Attr {
		if a.Key == "property" {
			if a.Val[:3] == "og:" {
				return true
			}
		}
	}
	return false
}
func getOgTagValues(n *html.Node) Tag {
	var (
		name  string
		value string
	)

	for _, a := range n.Attr {
		if a.Key == "property" {
			name = a.Val
		}
		if a.Key == "content" {
			value = a.Val
		}
	}

	return Tag{
		Name:  name,
		Value: value,
	}
}

func GetOgMetaTags(ctx context.Context, url string) ([]Tag, error) {
	ctx, span := tracer.Start(ctx, "get-og-meta-tags")
	defer span.End()

	body, err := GetBody(ctx, url)
	if err != nil {
		return nil, err
	}

	return getOgTags(body)
}

func GetRequiredTags(ctx context.Context, tags []Tag) RequiredTags {
	var rt RequiredTags

	for _, t := range tags {
		switch t.Name {
		case OgTitle:
			rt.Title = t
		case OgType:
			rt.Type = t
		case OgImage:
			rt.Image = t
		case OgUrl:
			rt.Url = t
		}
	}

	return rt
}
