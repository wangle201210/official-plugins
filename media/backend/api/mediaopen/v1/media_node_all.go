// This file declares the public mediaopen full node-list endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListAllNodesReq defines the request for reading all media node configs without pagination.
type ListAllNodesReq struct {
	g.Meta      `path:"/nodes/all" method:"get" tags:"媒体配置" summary:"查询全量节点数据" dc:"查询全部媒体节点配置，不分页。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
}

// ListAllNodesRes defines the full media node-list response.
type ListAllNodesRes struct {
	List []*NodeInfo `json:"list" dc:"节点列表" eg:"[]"`
}

// NodeInfo defines one public media node row.
type NodeInfo struct {
	Id         int    `json:"id" dc:"ID" eg:"1"`
	NodeNum    int    `json:"nodeNum" dc:"节点编号" eg:"1"`
	Name       string `json:"name" dc:"节点名称" eg:"华东节点"`
	QnUrl      string `json:"qnUrl" dc:"节点网关地址" eg:"https://qn.example.com"`
	BasicUrl   string `json:"basicUrl" dc:"基础平台网关地址" eg:"https://basic.example.com"`
	DnUrl      string `json:"dnUrl" dc:"属地网关地址" eg:"https://dn.example.com"`
	CreatorId  int    `json:"creatorId" dc:"创建人ID" eg:"1"`
	CreateTime *int64 `json:"createTime" dc:"创建时间，Unix timestamp in milliseconds" eg:"1778733600000"`
	UpdaterId  int    `json:"updaterId" dc:"修改人ID" eg:"1"`
	UpdateTime *int64 `json:"updateTime" dc:"修改时间，Unix timestamp in milliseconds" eg:"1778733600000"`
}
