// admin_photo_audit.go defines protected activation-photo audit and repair APIs.
package v1

import "github.com/gogf/gf/v2/frame/g"

type ListPhotoAuditReq struct {
	g.Meta   `path:"/plugins/sicau-niu/admin/audit/photos" method:"get" tags:"Sicau Niu Admin" summary:"查询激活照片审计列表" dc:"Return a bounded page of photo-bearing activations with batch-assembled player and cattle display fields. Object paths are not exposed. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	UserId   int64 `json:"userId" dc:"Optional player filter" eg:"0"`
	NiuId    int64 `json:"niuId" dc:"Optional cattle filter" eg:"0"`
	PageNum  int   `json:"pageNum" dc:"Page number" eg:"1"`
	PageSize int   `json:"pageSize" dc:"Page size capped at 100" eg:"20"`
}

type ListPhotoAuditRes struct {
	List  []*PhotoAuditItem `json:"list" dc:"Photo audit rows" eg:"[]"`
	Total int               `json:"total" dc:"Matched row count" eg:"0"`
}

type PhotoAuditItem struct {
	ActivationId int64  `json:"activationId" dc:"Activation ID" eg:"1"`
	PhotoId      string `json:"photoId" dc:"Opaque photo identifier" eg:"550e8400-e29b-41d4-a716-446655440000"`
	UserId       int64  `json:"userId" dc:"Player ID" eg:"2"`
	Nickname     string `json:"nickname" dc:"Player nickname" eg:"川农同学"`
	NiuId        int64  `json:"niuId" dc:"Cattle ID" eg:"7"`
	NiuName      string `json:"niuName" dc:"Cattle name" eg:"信息工程学院牛"`
	NiuCode      string `json:"niuCode" dc:"Cattle code" eg:"N007"`
	ContentType  string `json:"contentType" dc:"Standardized MIME type" eg:"image/webp"`
	SizeBytes    int64  `json:"sizeBytes" dc:"Stored byte size" eg:"204800"`
	ActivatedAt  *int64 `json:"activatedAt" dc:"Activation time as Unix milliseconds" eg:"1776333600000"`
}

type AuditPhotoContentReq struct {
	g.Meta  `path:"/plugins/sicau-niu/admin/audit/photos/{photoId}" method:"get" tags:"Sicau Niu Admin" summary:"读取激活照片审计内容" dc:"Return base64 photo evidence through the protected operator route without exposing object storage paths. Protected by host unified permission check." permission:"sicau-niu:record:list"`
	PhotoId string `json:"photoId" v:"required|max-length:64" dc:"Opaque photo identifier" eg:"550e8400-e29b-41d4-a716-446655440000"`
}

type AuditPhotoContentRes struct {
	ContentType string `json:"contentType" dc:"Image MIME type" eg:"image/webp"`
	SizeBytes   int64  `json:"sizeBytes" dc:"Image byte size" eg:"204800"`
	ImageBase64 string `json:"imageBase64" dc:"Base64 encoded image bytes" eg:"UklGRg..."`
}

type RevokeActivationReq struct {
	g.Meta `path:"/plugins/sicau-niu/admin/audit/activations/{id}" method:"delete" tags:"Sicau Niu Admin" summary:"撤销错误激活" dc:"Soft-delete one erroneous activation while retaining photo evidence and repairing first-activator state. Protected by host unified permission check." permission:"sicau-niu:record:revoke"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Activation ID" eg:"1"`
}

type RevokeActivationRes struct{}
