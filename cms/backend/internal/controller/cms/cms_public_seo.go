// This file implements the CMS public SEO endpoints: sitemap.xml, rss.xml,
// and robots.txt. XML documents are marshaled with encoding/xml so all user
// content is escaped, and article links reuse the public site URL shapes.

package cms

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// publicSeoSitemapArticleLimit groups the fixed output bounds and protocol
// constants used by the public SEO endpoints.
const (
	publicSeoSitemapArticleLimit = 5000
	publicSeoRssArticleLimit     = 50
	publicSeoSitemapXmlns        = "http://www.sitemaps.org/schemas/sitemap/0.9"
	publicSeoRssVersion          = "2.0"
	publicSeoXMLContentType      = "application/xml; charset=utf-8"
	publicSeoTextContentType     = "text/plain; charset=utf-8"
	publicSeoSitemapPath         = "/cms-site/sitemap.xml"
)

// publicSeoSitemapURL is one sitemap <url> entry.
type publicSeoSitemapURL struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

// publicSeoSitemapURLSet is the sitemap <urlset> document root.
type publicSeoSitemapURLSet struct {
	XMLName xml.Name              `xml:"urlset"`
	Xmlns   string                `xml:"xmlns,attr"`
	URLs    []publicSeoSitemapURL `xml:"url"`
}

// publicSeoRssItem is one RSS 2.0 <item> entry.
type publicSeoRssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

// publicSeoRssChannel is the RSS 2.0 <channel> element.
type publicSeoRssChannel struct {
	Title       string             `xml:"title"`
	Link        string             `xml:"link"`
	Description string             `xml:"description"`
	Items       []publicSeoRssItem `xml:"item"`
}

// publicSeoRssDocument is the RSS 2.0 document root.
type publicSeoRssDocument struct {
	XMLName xml.Name            `xml:"rss"`
	Version string              `xml:"version,attr"`
	Channel publicSeoRssChannel `xml:"channel"`
}

// PublicFrontendSitemap serves the public sitemap.xml document.
func (c *ControllerV1) PublicFrontendSitemap(r *ghttp.Request) {
	if r == nil {
		return
	}
	content, err := c.cmsSvc.GetPublicSeoContent(r.GetCtx(), publicSeoSitemapArticleLimit)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS sitemap is temporarily unavailable.")
		return
	}
	base := publicSeoBaseURL(content.Site)
	urls := []publicSeoSitemapURL{{Loc: publicSeoAbsoluteURL(base, "/cms-site")}}
	for _, category := range content.Categories {
		href := publicFrontendCategoryPathHref(category.Path)
		if href == "" {
			continue
		}
		urls = append(urls, publicSeoSitemapURL{Loc: publicSeoAbsoluteURL(base, href), Lastmod: publicSeoDate(category.UpdatedAt)})
	}
	for _, article := range content.Articles {
		urls = append(urls, publicSeoSitemapURL{Loc: publicSeoAbsoluteURL(base, publicSeoArticleHref(article.Slug)), Lastmod: publicSeoDate(publicSeoLatestTime(article.PublishedAt, article.UpdatedAt))})
	}
	for _, product := range content.Products {
		urls = append(urls, publicSeoSitemapURL{Loc: publicSeoAbsoluteURL(base, publicSeoProductHref(product.Slug)), Lastmod: publicSeoDate(publicSeoLatestTime(product.PublishedAt, product.UpdatedAt))})
	}
	writePublicSeoXML(r, publicSeoSitemapURLSet{Xmlns: publicSeoSitemapXmlns, URLs: urls})
}

// PublicFrontendRss serves the public rss.xml feed of the newest articles.
func (c *ControllerV1) PublicFrontendRss(r *ghttp.Request) {
	if r == nil {
		return
	}
	content, err := c.cmsSvc.GetPublicSeoContent(r.GetCtx(), publicSeoRssArticleLimit)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS RSS feed is temporarily unavailable.")
		return
	}
	base := publicSeoBaseURL(content.Site)
	channel := publicSeoRssChannel{
		Title:       content.Site.Name,
		Link:        publicSeoAbsoluteURL(base, "/cms-site"),
		Description: publicSeoChannelDescription(content.Site),
		Items:       make([]publicSeoRssItem, 0, len(content.Articles)),
	}
	for _, article := range content.Articles {
		link := publicSeoAbsoluteURL(base, publicSeoArticleHref(article.Slug))
		channel.Items = append(channel.Items, publicSeoRssItem{
			Title:       article.Title,
			Link:        link,
			Description: publicFrontendPlainText(article.Summary),
			PubDate:     publicSeoRFC1123(article.PublishedAt),
			GUID:        link,
		})
	}
	writePublicSeoXML(r, publicSeoRssDocument{Version: publicSeoRssVersion, Channel: channel})
}

// PublicFrontendRobots serves the public robots.txt with a sitemap pointer.
func (c *ControllerV1) PublicFrontendRobots(r *ghttp.Request) {
	if r == nil {
		return
	}
	content, err := c.cmsSvc.GetPublicSeoContent(r.GetCtx(), 1)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS robots.txt is temporarily unavailable.")
		return
	}
	base := publicSeoBaseURL(content.Site)
	var builder strings.Builder
	builder.WriteString("User-agent: *\n")
	builder.WriteString("Allow: /cms-site\n")
	builder.WriteString("Sitemap: " + publicSeoAbsoluteURL(base, publicSeoSitemapPath) + "\n")
	r.Response.Header().Set("Content-Type", publicSeoTextContentType)
	r.Response.Write(builder.String())
	r.ExitAll()
}

// writePublicSeoXML marshals one document and writes it with the XML header.
func writePublicSeoXML(r *ghttp.Request, document any) {
	payload, err := xml.Marshal(document)
	if err != nil {
		writePublicFrontendStatus(r, http.StatusInternalServerError, "CMS XML document could not be rendered.")
		return
	}
	r.Response.Header().Set("Content-Type", publicSeoXMLContentType)
	r.Response.Write(xml.Header)
	r.Response.Write(payload)
	r.ExitAll()
}

// publicSeoBaseURL extracts the configured absolute site origin; an empty
// result keeps SEO links relative.
func publicSeoBaseURL(site *cmssvc.SiteItem) string {
	if site == nil {
		return ""
	}
	domain := strings.TrimRight(strings.TrimSpace(site.Domain), "/")
	if domain == "" {
		return ""
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}
	return domain
}

// publicSeoAbsoluteURL joins the configured origin with a public path,
// falling back to the relative path when no domain is configured.
func publicSeoAbsoluteURL(base string, path string) string {
	if base == "" {
		return path
	}
	return base + path
}

// publicSeoArticleHref builds the public article detail path for one slug.
func publicSeoArticleHref(slug string) string {
	values := url.Values{}
	values.Set("article", slug)
	return "/cms-site?" + values.Encode()
}

// publicSeoProductHref builds the public product detail path for one slug.
func publicSeoProductHref(slug string) string {
	values := url.Values{}
	values.Set("product", slug)
	return "/cms-site?" + values.Encode()
}

// publicSeoChannelDescription picks the best available site description text.
func publicSeoChannelDescription(site *cmssvc.SiteItem) string {
	if site == nil {
		return ""
	}
	if description := strings.TrimSpace(site.Description); description != "" {
		return description
	}
	return strings.TrimSpace(site.Slogan)
}

// publicSeoDate formats a time as a W3C sitemap lastmod date.
func publicSeoDate(value *gtime.Time) string {
	if value == nil {
		return ""
	}
	return value.Layout("2006-01-02")
}

// publicSeoRFC1123 formats a time as an RSS pubDate.
func publicSeoRFC1123(value *gtime.Time) string {
	if value == nil {
		return ""
	}
	return value.Time.Format(time.RFC1123Z)
}

// publicSeoLatestTime returns the later of two optional times for lastmod.
func publicSeoLatestTime(first *gtime.Time, second *gtime.Time) *gtime.Time {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	if second.After(first) {
		return second
	}
	return first
}
