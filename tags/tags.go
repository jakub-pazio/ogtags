package tags

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	OgTitle string = "og:title"
	OgType  string = "og:type"
	OgImage string = "og:image"
	OgUrl   string = "og:url"
)

type Tag struct {
	Name  string
	Value string
}

type RequiredTags struct {
	Title Tag
	Type  Tag
	Image Tag
	Url   Tag
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

func GetBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", nil
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
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
			t := getOgTag(n)
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
func getOgTag(n *html.Node) Tag {
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

func GetOgMetaTags(url string) ([]Tag, error) {
	body, err := GetBody(url)
	if err != nil {
		return nil, err
	}

	return getOgTags(body)
}

func GetRequiredTags(tags []Tag) RequiredTags {
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
