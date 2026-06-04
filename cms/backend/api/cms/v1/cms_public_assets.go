// This file declares public CMS links and slides APIs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicLinkListReq defines the public request for enabled friendly links.
type PublicLinkListReq struct {
	g.Meta `path:"/cms/public/links" method:"get" tags:"CMS Public" summary:"Get public CMS links" dc:"Get enabled friendly links without management authentication."`
}

// PublicLinkListRes defines the public response for enabled friendly links.
type PublicLinkListRes struct {
	List []*LinkItem `json:"list" dc:"Enabled friendly links" eg:"[]"`
}

// PublicSlideListReq defines the public request for enabled slides.
type PublicSlideListReq struct {
	g.Meta `path:"/cms/public/slides" method:"get" tags:"CMS Public" summary:"Get public CMS slides" dc:"Get enabled slides without management authentication."`
}

// PublicSlideListRes defines the public response for enabled slides.
type PublicSlideListRes struct {
	List []*SlideItem `json:"list" dc:"Enabled slides" eg:"[]"`
}
