// This file declares the CMS sample data loading API.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SiteLoadSampleDataReq defines the request for loading packaged CMS sample data.
type SiteLoadSampleDataReq struct {
	g.Meta `path:"/cms/site/sample-data" method:"post" tags:"CMS Site" summary:"Load CMS sample data" dc:"Clear current CMS business content and load the packaged starter site dataset." permission:"cms:site:sample"`
}

// SiteLoadSampleDataRes defines the response for loading packaged CMS sample data.
type SiteLoadSampleDataRes struct{}
