package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type UploadPhotoReq struct {
	g.Meta    `path:"/plugins/sicau-niu/player/photos" method:"post" mime:"multipart/form-data" tags:"寻牛小程序" summary:"上传激活照片" dc:"Upload one private activation photo from multipart field file. JPEG, PNG and HEIC images up to 5 MiB are transcoded to WebP no larger than 300 KiB. Ten successful uploads are allowed per Beijing natural day. Reusing requestId returns the first photoId. Requires a valid player token."`
	File      *ghttp.UploadFile `json:"file" type:"file" v:"required" dc:"Required activation photo file field; JPEG, PNG or HEIC, up to 5 MiB" eg:"activation.jpg"`
	RequestId string            `json:"requestId" form:"requestId" v:"required|max-length:64" dc:"Player-scoped idempotency key" eg:"photo-1b2a3c4d"`
}

type UploadPhotoRes struct {
	PhotoId     string `json:"photoId" dc:"Opaque activation photo identifier" eg:"550e8400-e29b-41d4-a716-446655440000"`
	ContentType string `json:"contentType" dc:"Standardized image MIME type" eg:"image/webp"`
	SizeBytes   int64  `json:"sizeBytes" dc:"Standardized image size in bytes, no more than 307200" eg:"204800"`
}

type PhotoContentReq struct {
	g.Meta  `path:"/plugins/sicau-niu/player/photos/{photoId}" method:"get" tags:"寻牛小程序" summary:"读取本人激活照片" dc:"Return a private activation photo only when it belongs to the current player. Missing and foreign photos use the same not-found response. Requires a valid player token."`
	PhotoId string `json:"photoId" v:"required" dc:"Opaque photo identifier returned by upload" eg:"550e8400-e29b-41d4-a716-446655440000"`
}

type PhotoContentRes struct {
	ContentType string `json:"contentType" dc:"Detected image MIME type" eg:"image/jpeg"`
	SizeBytes   int64  `json:"sizeBytes" dc:"Image size in bytes" eg:"204800"`
	ImageBase64 string `json:"imageBase64" dc:"Base64 encoded image bytes" eg:"UklGRg..."`
}
