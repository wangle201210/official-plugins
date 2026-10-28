// This file declares the public mediaopen full node-list endpoint.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListAllNodesReq defines the request for reading all media node configs without pagination.
type ListAllNodesReq struct {
	g.Meta      `path:"/nodes/all" method:"get" tags:"内部接口" summary:"查询全量节点配置" dc:"读取节点管理配置表的全部节点，返回节点编号、名称、网关地址及维护信息，不是采集上报节点总览。无需分页或筛选参数，按nodeNum升序；编号0至255且唯一，正常最多256条，全部返回。沿用内部API Key鉴权，列表位于统一响应data.list，无记录时为空数组。" access:"public"`
	InnerApiKey string `json:"X-Inner-Api-Key" in:"header" dc:"内部接口API Key；默认值media；显式配置innerapi.apiKey为空时可不传" eg:"media"`
}

// ListAllNodesRes defines the full media node-list response.
type ListAllNodesRes struct {
	List []*NodeInfo `json:"list" dc:"全量节点配置，按nodeNum升序，包含节点管理全部字段；空结果为[]，不分页" eg:"[]"`
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
	CreateTime *int64 `json:"createTime" nullable:"true" format:"int64" dc:"创建时间，Unix timestamp in milliseconds；缺失时为null" eg:"1778733600000"`
	UpdaterId  int    `json:"updaterId" dc:"修改人ID" eg:"1"`
	UpdateTime *int64 `json:"updateTime" nullable:"true" format:"int64" dc:"修改时间，Unix timestamp in milliseconds；缺失时为null" eg:"1778733600000"`
}
