// This file declares the CMS site data clearing API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SiteClearDataReq defines the request for clearing CMS business content.
type SiteClearDataReq struct {
	g.Meta `path:"/cms/site/data" method:"delete" tags:"CMS Site" summary:"Clear CMS site data" dc:"Clear CMS categories, articles, tags, slides, links, messages, and reset the default site settings." permission:"cms:site:purge"`
}

// SiteClearDataRes defines the response for clearing CMS business content.
type SiteClearDataRes struct{}
