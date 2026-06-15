// This file verifies the CMS public SEO document builders: URL shapes, date
// formats, and XML escaping of user-controlled text.

package cms

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// TestPublicSeoBaseURLNormalizesDomains verifies domain trimming, scheme
// defaults, and the relative fallback when no domain is configured.
func TestPublicSeoBaseURLNormalizesDomains(t *testing.T) {
	cases := []struct {
		name   string
		domain string
		want   string
	}{
		{name: "empty domain keeps links relative", domain: "  ", want: ""},
		{name: "trailing slash is trimmed", domain: "https://cms.example.com/", want: "https://cms.example.com"},
		{name: "bare host gains https scheme", domain: "cms.example.com", want: "https://cms.example.com"},
		{name: "http scheme is preserved", domain: "http://cms.example.com", want: "http://cms.example.com"},
	}
	for _, testCase := range cases {
		site := &cmssvc.SiteItem{}
		site.Domain = testCase.domain
		if got := publicSeoBaseURL(site); got != testCase.want {
			t.Fatalf("%s: expected %q, got %q", testCase.name, testCase.want, got)
		}
	}
	if got := publicSeoBaseURL(nil); got != "" {
		t.Fatalf("expected empty base URL for nil site, got %q", got)
	}
	if got := publicSeoAbsoluteURL("", "/cms-site"); got != "/cms-site" {
		t.Fatalf("expected relative fallback, got %q", got)
	}
	if got := publicSeoAbsoluteURL("https://cms.example.com", "/cms-site"); got != "https://cms.example.com/cms-site" {
		t.Fatalf("expected absolute URL, got %q", got)
	}
}

// TestPublicSeoArticleHrefEscapesSlugs verifies article links keep the public
// detail URL shape with query escaping.
func TestPublicSeoArticleHrefEscapesSlugs(t *testing.T) {
	if got := publicSeoArticleHref("hello-world"); got != "/cms-site?article=hello-world" {
		t.Fatalf("expected plain article href, got %q", got)
	}
	if got := publicSeoArticleHref("a&b c"); got != "/cms-site?article=a%26b+c" {
		t.Fatalf("expected escaped article href, got %q", got)
	}
}

// TestPublicSeoSitemapMarshalEscapesContent verifies sitemap documents stay
// parseable when locations contain XML special characters.
func TestPublicSeoSitemapMarshalEscapesContent(t *testing.T) {
	document := publicSeoSitemapURLSet{
		Xmlns: publicSeoSitemapXmlns,
		URLs: []publicSeoSitemapURL{
			{Loc: "/cms-site?article=a%26b", Lastmod: "2026-06-12"},
			{Loc: "/cms-site/news<script>/"},
		},
	}
	payload, err := xml.Marshal(document)
	if err != nil {
		t.Fatalf("marshal sitemap document: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "<script>") {
		t.Fatalf("expected escaped sitemap content, got %s", text)
	}
	var parsed publicSeoSitemapURLSet
	if err = xml.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("re-parse sitemap document: %v", err)
	}
	if len(parsed.URLs) != 2 || parsed.URLs[0].Loc != "/cms-site?article=a%26b" {
		t.Fatalf("expected round-tripped sitemap URLs, got %+v", parsed.URLs)
	}
}

// TestPublicSeoRssMarshalStripsAndEscapesRichText verifies RSS descriptions
// drop HTML markup and survive an XML round trip.
func TestPublicSeoRssMarshalStripsAndEscapesRichText(t *testing.T) {
	summary := publicFrontendPlainText(`<p>Hello <b>world</b> &amp; "friends"</p>`)
	document := publicSeoRssDocument{
		Version: publicSeoRssVersion,
		Channel: publicSeoRssChannel{
			Title:       `Site <Name> & Co`,
			Link:        "/cms-site",
			Description: "demo",
			Items:       []publicSeoRssItem{{Title: "T", Link: "/cms-site?article=t", Description: summary, GUID: "/cms-site?article=t"}},
		},
	}
	payload, err := xml.Marshal(document)
	if err != nil {
		t.Fatalf("marshal RSS document: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "<b>") || strings.Contains(text, "<p>") {
		t.Fatalf("expected stripped rich text in RSS description, got %s", text)
	}
	var parsed publicSeoRssDocument
	if err = xml.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("re-parse RSS document: %v", err)
	}
	if parsed.Channel.Title != `Site <Name> & Co` {
		t.Fatalf("expected round-tripped channel title, got %q", parsed.Channel.Title)
	}
	if parsed.Channel.Items[0].Description != `Hello world & "friends"` {
		t.Fatalf("expected plain-text description, got %q", parsed.Channel.Items[0].Description)
	}
}

// TestPublicSeoTimeFormatting verifies lastmod and pubDate formats plus the
// latest-time selection used for sitemap entries.
func TestPublicSeoTimeFormatting(t *testing.T) {
	moment := gtime.New(time.Date(2026, 6, 12, 8, 30, 0, 0, time.UTC))
	if got := publicSeoDate(moment); got != "2026-06-12" {
		t.Fatalf("expected W3C date, got %q", got)
	}
	if got := publicSeoDate(nil); got != "" {
		t.Fatalf("expected empty date for nil time, got %q", got)
	}
	if got := publicSeoRFC1123(moment); !strings.Contains(got, "12 Jun 2026") {
		t.Fatalf("expected RFC1123 pubDate, got %q", got)
	}
	if got := publicSeoRFC1123(nil); got != "" {
		t.Fatalf("expected empty pubDate for nil time, got %q", got)
	}

	earlier := gtime.New(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if got := publicSeoLatestTime(earlier, moment); !got.Equal(moment) {
		t.Fatalf("expected later time selected, got %v", got)
	}
	if got := publicSeoLatestTime(nil, moment); !got.Equal(moment) {
		t.Fatalf("expected second time when first is nil, got %v", got)
	}
	if got := publicSeoLatestTime(moment, nil); !got.Equal(moment) {
		t.Fatalf("expected first time when second is nil, got %v", got)
	}
}
