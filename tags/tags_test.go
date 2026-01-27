package tags

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const htmlbody = `
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

func TestGetOgTags(t *testing.T) {
	got, err := getOgTags(htmlbody)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := []Tag{
		{Name: "og:title", Value: "The Rock"},
		{Name: "og:type", Value: "video.movie"},
		{Name: "og:url", Value: "https://www.imdb.com/title/tt0117500/"},
		{Name: "og:image", Value: "https://ia.media-imdb.com/images/rock.jpg"},
	}

	tagSortFunc := func(a, b Tag) int {
		return strings.Compare(a.Name, b.Name)
	}

	slices.SortFunc(got, tagSortFunc)
	slices.SortFunc(want, tagSortFunc)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("getOgTags mismatch (-want +got):\n%s", diff)
	}
}
