// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsSite is the golang structure for table cms_site.
type CmsSite struct {
	Id           int64       `json:"id"           orm:"id"            description:"Site ID"`
	SiteKey      string      `json:"siteKey"      orm:"site_key"      description:"Stable site key"`
	Name         string      `json:"name"         orm:"name"          description:"Site name"`
	Logo         string      `json:"logo"         orm:"logo"          description:"Site logo URL"`
	Weixin       string      `json:"weixin"       orm:"weixin"        description:"WeChat QR code image URL"`
	Domain       string      `json:"domain"       orm:"domain"        description:"Primary site domain"`
	Slogan       string      `json:"slogan"       orm:"slogan"        description:"Site slogan"`
	Keywords     string      `json:"keywords"     orm:"keywords"      description:"SEO keywords"`
	Description  string      `json:"description"  orm:"description"   description:"SEO description"`
	Icp          string      `json:"icp"          orm:"icp"           description:"ICP record number"`
	Contact      string      `json:"contact"      orm:"contact"       description:"Contact person"`
	Phone        string      `json:"phone"        orm:"phone"         description:"Contact phone"`
	Email        string      `json:"email"        orm:"email"         description:"Contact email"`
	Address      string      `json:"address"      orm:"address"       description:"Contact address"`
	Status       int         `json:"status"       orm:"status"        description:"Status: 0=disabled, 1=enabled"`
	CreatedBy    int64       `json:"createdBy"    orm:"created_by"    description:"Creator user ID"`
	UpdatedBy    int64       `json:"updatedBy"    orm:"updated_by"    description:"Updater user ID"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"Creation time"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"Update time"`
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:"Deletion time"`
	ShowMessages int         `json:"showMessages" orm:"show_messages" description:"Show approved visitor messages on public message page: 0=no, 1=yes"`
}
