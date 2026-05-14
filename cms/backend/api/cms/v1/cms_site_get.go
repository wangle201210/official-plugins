// This file declares the CMS site-settings detail API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SiteGetReq defines the request for reading CMS site settings.
type SiteGetReq struct {
	g.Meta `path:"/cms/site" method:"get" tags:"CMS Site" summary:"Get CMS site settings" dc:"Get the editable CMS site settings." permission:"cms:workspace:view"`
}

// SiteGetRes defines the response for reading CMS site settings.
type SiteGetRes struct {
	*SiteItem
}
