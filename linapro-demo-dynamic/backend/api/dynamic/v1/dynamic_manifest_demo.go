// This file defines the manifest host-service demo DTOs for the dynamic
// plugin sample.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ManifestDemoReq is the request for querying the manifest host-service demo.
type ManifestDemoReq struct {
	g.Meta `path:"/manifest-demo" method:"get" tags:"Dynamic Plugin Demo" summary:"Query the manifest host service demonstration" dc:"Return values read from the two packaged manifest resources explicitly authorized in plugin.yaml: config/profile.yaml and config/config.yaml. This read-only endpoint demonstrates the flow from service: manifest declaration to manifest.get use without triggering storage, data, or network side effects." access:"login" permission:"linapro-demo-dynamic:backend:view" operLog:"other"`
}

// ManifestDemoRes is the response for querying the manifest host-service demo.
type ManifestDemoRes = HostCallDemoManifestRes
