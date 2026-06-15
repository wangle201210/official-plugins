// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CmsProduct is the golang structure for table cms_product.
type CmsProduct struct {
	Id          int64       `json:"id"          orm:"id"           description:"Product ID"`
	CategoryId  int64       `json:"categoryId"  orm:"category_id"  description:"Category ID"`
	Name        string      `json:"name"        orm:"name"         description:"Product name"`
	Slug        string      `json:"slug"        orm:"slug"         description:"Public URL slug"`
	Summary     string      `json:"summary"     orm:"summary"      description:"Product summary"`
	Cover       string      `json:"cover"       orm:"cover"        description:"Cover image URL"`
	Gallery     string      `json:"gallery"     orm:"gallery"      description:"Gallery image URL list as JSON array text"`
	Price       string      `json:"price"       orm:"price"        description:"Display price text"`
	Spec        string      `json:"spec"        orm:"spec"         description:"Specification summary"`
	Content     string      `json:"content"     orm:"content"      description:"Product detail HTML"`
	Keywords    string      `json:"keywords"    orm:"keywords"     description:"SEO keywords"`
	Description string      `json:"description" orm:"description"  description:"SEO description"`
	Sort        int         `json:"sort"        orm:"sort"         description:"Display order"`
	Status      int         `json:"status"      orm:"status"       description:"Status: 0=draft, 1=published"`
	IsTop       int         `json:"isTop"       orm:"is_top"       description:"Top flag: 0=no, 1=yes"`
	IsRecommend int         `json:"isRecommend" orm:"is_recommend" description:"Recommend flag: 0=no, 1=yes"`
	Views       int64       `json:"views"       orm:"views"        description:"View count"`
	PublishedAt *gtime.Time `json:"publishedAt" orm:"published_at" description:"Publication time"`
	CreatedBy   int64       `json:"createdBy"   orm:"created_by"   description:"Creator user ID"`
	UpdatedBy   int64       `json:"updatedBy"   orm:"updated_by"   description:"Updater user ID"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:"Creation time"`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:"Update time"`
	DeletedAt   *gtime.Time `json:"deletedAt"   orm:"deleted_at"   description:"Deletion time"`
}
