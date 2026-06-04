// This file declares the public CMS site API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// PublicSiteReq defines the public request for reading enabled site settings.
type PublicSiteReq struct {
	g.Meta `path:"/cms/public/site" method:"get" tags:"CMS Public" summary:"Get public CMS site settings" dc:"Get public CMS site settings without management authentication."`
}

// PublicSiteRes defines the public response for reading enabled site settings.
type PublicSiteRes struct {
	*SiteItem
}
