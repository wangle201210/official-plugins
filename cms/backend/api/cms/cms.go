// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

type ICmsV1 interface {
	AlbumList(ctx context.Context, req *v1.AlbumListReq) (res *v1.AlbumListRes, err error)
	AlbumGet(ctx context.Context, req *v1.AlbumGetReq) (res *v1.AlbumGetRes, err error)
	AlbumCreate(ctx context.Context, req *v1.AlbumCreateReq) (res *v1.AlbumCreateRes, err error)
	AlbumUpdate(ctx context.Context, req *v1.AlbumUpdateReq) (res *v1.AlbumUpdateRes, err error)
	AlbumDelete(ctx context.Context, req *v1.AlbumDeleteReq) (res *v1.AlbumDeleteRes, err error)
	ArticleBatchDelete(ctx context.Context, req *v1.ArticleBatchDeleteReq) (res *v1.ArticleBatchDeleteRes, err error)
	ArticleBatchUpdate(ctx context.Context, req *v1.ArticleBatchUpdateReq) (res *v1.ArticleBatchUpdateRes, err error)
	ArticleCreate(ctx context.Context, req *v1.ArticleCreateReq) (res *v1.ArticleCreateRes, err error)
	ArticleDelete(ctx context.Context, req *v1.ArticleDeleteReq) (res *v1.ArticleDeleteRes, err error)
	ArticleGet(ctx context.Context, req *v1.ArticleGetReq) (res *v1.ArticleGetRes, err error)
	ArticleList(ctx context.Context, req *v1.ArticleListReq) (res *v1.ArticleListRes, err error)
	ArticleUpdate(ctx context.Context, req *v1.ArticleUpdateReq) (res *v1.ArticleUpdateRes, err error)
	CategoryCreate(ctx context.Context, req *v1.CategoryCreateReq) (res *v1.CategoryCreateRes, err error)
	CategoryDelete(ctx context.Context, req *v1.CategoryDeleteReq) (res *v1.CategoryDeleteRes, err error)
	CategoryList(ctx context.Context, req *v1.CategoryListReq) (res *v1.CategoryListRes, err error)
	CategoryUpdate(ctx context.Context, req *v1.CategoryUpdateReq) (res *v1.CategoryUpdateRes, err error)
	LinkCreate(ctx context.Context, req *v1.LinkCreateReq) (res *v1.LinkCreateRes, err error)
	LinkDelete(ctx context.Context, req *v1.LinkDeleteReq) (res *v1.LinkDeleteRes, err error)
	LinkList(ctx context.Context, req *v1.LinkListReq) (res *v1.LinkListRes, err error)
	LinkUpdate(ctx context.Context, req *v1.LinkUpdateReq) (res *v1.LinkUpdateRes, err error)
	MessageDelete(ctx context.Context, req *v1.MessageDeleteReq) (res *v1.MessageDeleteRes, err error)
	MessageList(ctx context.Context, req *v1.MessageListReq) (res *v1.MessageListRes, err error)
	MessageUpdate(ctx context.Context, req *v1.MessageUpdateReq) (res *v1.MessageUpdateRes, err error)
	ProductCreate(ctx context.Context, req *v1.ProductCreateReq) (res *v1.ProductCreateRes, err error)
	ProductGet(ctx context.Context, req *v1.ProductGetReq) (res *v1.ProductGetRes, err error)
	ProductDelete(ctx context.Context, req *v1.ProductDeleteReq) (res *v1.ProductDeleteRes, err error)
	ProductList(ctx context.Context, req *v1.ProductListReq) (res *v1.ProductListRes, err error)
	ProductUpdate(ctx context.Context, req *v1.ProductUpdateReq) (res *v1.ProductUpdateRes, err error)
	PublicAlbumList(ctx context.Context, req *v1.PublicAlbumListReq) (res *v1.PublicAlbumListRes, err error)
	PublicAlbumGet(ctx context.Context, req *v1.PublicAlbumGetReq) (res *v1.PublicAlbumGetRes, err error)
	PublicArticleList(ctx context.Context, req *v1.PublicArticleListReq) (res *v1.PublicArticleListRes, err error)
	PublicArticleGet(ctx context.Context, req *v1.PublicArticleGetReq) (res *v1.PublicArticleGetRes, err error)
	PublicLinkList(ctx context.Context, req *v1.PublicLinkListReq) (res *v1.PublicLinkListRes, err error)
	PublicSlideList(ctx context.Context, req *v1.PublicSlideListReq) (res *v1.PublicSlideListRes, err error)
	PublicCategoryList(ctx context.Context, req *v1.PublicCategoryListReq) (res *v1.PublicCategoryListRes, err error)
	PublicMessageCreate(ctx context.Context, req *v1.PublicMessageCreateReq) (res *v1.PublicMessageCreateRes, err error)
	PublicMessageList(ctx context.Context, req *v1.PublicMessageListReq) (res *v1.PublicMessageListRes, err error)
	PublicProductList(ctx context.Context, req *v1.PublicProductListReq) (res *v1.PublicProductListRes, err error)
	PublicProductGet(ctx context.Context, req *v1.PublicProductGetReq) (res *v1.PublicProductGetRes, err error)
	PublicSite(ctx context.Context, req *v1.PublicSiteReq) (res *v1.PublicSiteRes, err error)
	SiteClearData(ctx context.Context, req *v1.SiteClearDataReq) (res *v1.SiteClearDataRes, err error)
	SiteGet(ctx context.Context, req *v1.SiteGetReq) (res *v1.SiteGetRes, err error)
	SiteLoadSampleData(ctx context.Context, req *v1.SiteLoadSampleDataReq) (res *v1.SiteLoadSampleDataRes, err error)
	SiteUpdate(ctx context.Context, req *v1.SiteUpdateReq) (res *v1.SiteUpdateRes, err error)
	SlideCreate(ctx context.Context, req *v1.SlideCreateReq) (res *v1.SlideCreateRes, err error)
	SlideDelete(ctx context.Context, req *v1.SlideDeleteReq) (res *v1.SlideDeleteRes, err error)
	SlideList(ctx context.Context, req *v1.SlideListReq) (res *v1.SlideListRes, err error)
	SlideUpdate(ctx context.Context, req *v1.SlideUpdateReq) (res *v1.SlideUpdateRes, err error)
}
