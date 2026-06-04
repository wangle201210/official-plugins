// This file declares the CMS site-settings update API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SiteUpdateReq defines the request for updating CMS site settings.
type SiteUpdateReq struct {
	g.Meta       `path:"/cms/site" method:"put" tags:"CMS Site" summary:"Update CMS site settings" dc:"Update editable CMS site settings." permission:"cms:site:edit"`
	Name         string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Site name" eg:"LinaPro CMS"`
	Logo         string `json:"logo" dc:"Site logo URL" eg:"/uploads/logo.png"`
	Weixin       string `json:"weixin" dc:"WeChat QR code image URL" eg:"/uploads/weixin.png"`
	Domain       string `json:"domain" dc:"Primary site domain" eg:"https://example.com"`
	Slogan       string `json:"slogan" dc:"Site slogan" eg:"AI-native content delivery"`
	Keywords     string `json:"keywords" dc:"SEO keywords" eg:"LinaPro,CMS"`
	Description  string `json:"description" dc:"SEO description" eg:"LinaPro CMS site"`
	Icp          string `json:"icp" dc:"ICP record number" eg:"ICP 00000000"`
	Contact      string `json:"contact" dc:"Contact person" eg:"Admin"`
	Phone        string `json:"phone" dc:"Contact phone" eg:"13800000000"`
	Email        string `json:"email" dc:"Contact email" eg:"hello@example.com"`
	Address      string `json:"address" dc:"Contact address" eg:"Shanghai"`
	Status       int    `json:"status" v:"in:0,1#gf.gvalid.rule.in" dc:"Status: 0=disabled, 1=enabled" eg:"1"`
	ShowMessages int    `json:"showMessages" v:"in:0,1#gf.gvalid.rule.in" dc:"Show approved visitor messages on public message page: 0=no, 1=yes" eg:"0"`
}

// SiteUpdateRes defines the response for updating CMS site settings.
type SiteUpdateRes struct{}
