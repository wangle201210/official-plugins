// This file restores old GoAdmin generator routes against plugin-owned
// sys_tables, sys_columns, and sys_menu compatibility tables.

package uidentity

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

const (
	configKeyLegacyGenRoot      = "legacy.gen.outputRoot"
	configKeyLegacyGenFrontRoot = "legacy.gen.frontOutputRoot"

	defaultLegacyGenRoot      = "temp/uidentity-cas-gen/backend"
	defaultLegacyGenFrontRoot = "temp/uidentity-cas-gen/frontend"
)

const (
	legacyGenModelTemplate = `package models

import (
{{- if .HasTime }}
	"time"
{{- end }}

	"go-admin/common/models"
)

type {{.ClassName}} struct {
	models.Model
{{- range .ModelColumns }}
	{{.GoField}} {{.GoType}} ` + "`json:\"{{.JsonField}}\" gorm:\"type:{{.ColumnType}};comment:{{.Comment}}\"`" + `
{{- end }}
	models.ModelTime
	models.ControlBy
}

func ({{.ClassName}}) TableName() string {
	return "{{.TBName}}"
}

func (e *{{.ClassName}}) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *{{.ClassName}}) GetId() interface{} {
	return e.{{.PkGoField}}
}
`
	legacyGenAPITemplate = `package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/api"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	_ "github.com/go-admin-team/go-admin-core/sdk/pkg/response"

	"go-admin/app/{{.PackageName}}/models"
	"go-admin/app/{{.PackageName}}/service"
	"go-admin/app/{{.PackageName}}/service/dto"
	"go-admin/common/actions"
)

type {{.ClassName}} struct {
	api.Api
}

// GetPage 获取{{.TableComment}}列表
func (e {{.ClassName}}) GetPage(c *gin.Context) {
	req := dto.{{.ClassName}}GetPageReq{}
	s := service.{{.ClassName}}{}
	err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors
	if err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	list := make([]models.{{.ClassName}}, 0)
	var count int64
	if err = s.GetPage(&req, p, &list, &count); err != nil {
		e.Error(500, err, fmt.Sprintf("获取{{.TableComment}}失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.PageOK(list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
}

// Get 获取{{.TableComment}}
func (e {{.ClassName}}) Get(c *gin.Context) {
	req := dto.{{.ClassName}}GetReq{}
	s := service.{{.ClassName}}{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	var object models.{{.ClassName}}
	p := actions.GetPermissionFromContext(c)
	if err := s.Get(&req, p, &object); err != nil {
		e.Error(500, err, fmt.Sprintf("获取{{.TableComment}}失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(object, "查询成功")
}

// Insert 创建{{.TableComment}}
func (e {{.ClassName}}) Insert(c *gin.Context) {
	req := dto.{{.ClassName}}InsertReq{}
	s := service.{{.ClassName}}{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetCreateBy(user.GetUserId(c))
	if err := s.Insert(&req); err != nil {
		e.Error(500, err, fmt.Sprintf("创建{{.TableComment}}失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "创建成功")
}

// Update 修改{{.TableComment}}
func (e {{.ClassName}}) Update(c *gin.Context) {
	req := dto.{{.ClassName}}UpdateReq{}
	s := service.{{.ClassName}}{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	req.SetUpdateBy(user.GetUserId(c))
	p := actions.GetPermissionFromContext(c)
	if err := s.Update(&req, p); err != nil {
		e.Error(500, err, fmt.Sprintf("修改{{.TableComment}}失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "修改成功")
}

// Delete 删除{{.TableComment}}
func (e {{.ClassName}}) Delete(c *gin.Context) {
	s := service.{{.ClassName}}{}
	req := dto.{{.ClassName}}DeleteReq{}
	if err := e.MakeContext(c).MakeOrm().Bind(&req).MakeService(&s.Service).Errors; err != nil {
		e.Logger.Error(err)
		e.Error(500, err, err.Error())
		return
	}
	p := actions.GetPermissionFromContext(c)
	if err := s.Remove(&req, p); err != nil {
		e.Error(500, err, fmt.Sprintf("删除{{.TableComment}}失败，\r\n失败信息 %s", err.Error()))
		return
	}
	e.OK(req.GetId(), "删除成功")
}
`
	legacyGenJSTemplate = `import request from '@/utils/request'

// 查询{{.ClassName}}列表
export function list{{.ClassName}}(query) {
	return request({
		url: '/api/v1/{{.ModuleName}}',
		method: 'get',
		params: query
	})
}

// 查询{{.ClassName}}详细
export function get{{.ClassName}}({{.PkJsonField}}) {
	return request({
		url: '/api/v1/{{.ModuleName}}/' + {{.PkJsonField}},
		method: 'get'
	})
}

// 新增{{.ClassName}}
export function add{{.ClassName}}(data) {
	return request({
		url: '/api/v1/{{.ModuleName}}',
		method: 'post',
		data: data
	})
}

// 修改{{.ClassName}}
export function update{{.ClassName}}(data) {
	return request({
		url: '/api/v1/{{.ModuleName}}/' + data.{{.PkJsonField}},
		method: 'put',
		data: data
	})
}

// 删除{{.ClassName}}
export function del{{.ClassName}}(data) {
	return request({
		url: '/api/v1/{{.ModuleName}}',
		method: 'delete',
		data: data
	})
}
`
	legacyGenVueTemplate = `<template>
	<BasicLayout>
		<template #wrapper>
			<el-card class="box-card">
				<el-form ref="queryForm" :model="queryParams" :inline="true" label-width="68px">
{{- range .QueryColumns }}
					<el-form-item label="{{.ColumnComment}}" prop="{{.JsonField}}">
						<el-input v-model="queryParams.{{.JsonField}}" placeholder="请输入{{.ColumnComment}}" clearable size="small" @keyup.enter.native="handleQuery"/>
					</el-form-item>
{{- end }}
				</el-form>
				<el-table v-loading="loading" :data="{{.BusinessName}}List">
{{- range .ListColumns }}
					<el-table-column label="{{.ColumnComment}}" align="center" prop="{{.JsonField}}" :show-overflow-tooltip="true"/>
{{- end }}
				</el-table>
			</el-card>
		</template>
	</BasicLayout>
</template>
`
	legacyGenRouterTemplate = `package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/{{.PackageName}}/apis"
	"go-admin/common/actions"
	"go-admin/common/middleware"
)

func init() {
	routerCheckRole = append(routerCheckRole, register{{.ClassName}}Router)
}

func register{{.ClassName}}Router(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.{{.ClassName}}{}
	r := v1.Group("/{{.ModuleName}}").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	{
		r.GET("", actions.PermissionAction(), api.GetPage)
		r.GET("/:id", actions.PermissionAction(), api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", actions.PermissionAction(), api.Update)
		r.DELETE("", api.Delete)
	}
}
`
	legacyGenDTOTemplate = `package dto

import (
{{- if .HasTime }}
	"time"
{{- end }}

	"go-admin/app/{{.PackageName}}/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type {{.ClassName}}GetPageReq struct {
	dto.Pagination ` + "`search:\"-\"`" + `
{{- range .QueryColumns }}
	{{.GoField}} {{.GoType}} ` + "`form:\"{{.JsonField}}\" search:\"type:{{.SearchType}};column:{{.ColumnName}};table:{{$.TBName}}\" comment:\"{{.ColumnComment}}\"`" + `
{{- end }}
	{{.ClassName}}Order
}

type {{.ClassName}}Order struct {
{{- range .Columns }}
	{{.GoField}} string ` + "`form:\"{{.JsonField}}Order\" search:\"type:order;column:{{.ColumnName}};table:{{$.TBName}}\"`" + `
{{- end }}
}

func (m *{{.ClassName}}GetPageReq) GetNeedSearch() interface{} { return *m }

type {{.ClassName}}InsertReq struct {
{{- range .InsertColumns }}
	{{.GoField}} {{.GoType}} ` + "`json:\"{{.JsonField}}\" comment:\"{{.ColumnComment}}\"`" + `
{{- end }}
	common.ControlBy
}

func (s *{{.ClassName}}InsertReq) Generate(model *models.{{.ClassName}}) {
{{- range .InsertColumns }}
	model.{{.GoField}} = s.{{.GoField}}
{{- end }}
}

func (s *{{.ClassName}}InsertReq) GetId() interface{} { return s.{{.PkGoField}} }

type {{.ClassName}}UpdateReq struct {
	{{.PkGoField}} {{.PkGoType}} ` + "`uri:\"{{.PkJsonField}}\" comment:\"{{.PkColumnComment}}\"`" + `
{{- range .UpdateColumns }}
	{{.GoField}} {{.GoType}} ` + "`json:\"{{.JsonField}}\" comment:\"{{.ColumnComment}}\"`" + `
{{- end }}
	common.ControlBy
}

func (s *{{.ClassName}}UpdateReq) Generate(model *models.{{.ClassName}}) {
	model.Model = common.Model{Id: s.{{.PkGoField}}}
{{- range .UpdateColumns }}
	model.{{.GoField}} = s.{{.GoField}}
{{- end }}
}

func (s *{{.ClassName}}UpdateReq) GetId() interface{} { return s.{{.PkGoField}} }

type {{.ClassName}}GetReq struct {
	{{.PkGoField}} {{.PkGoType}} ` + "`uri:\"{{.PkJsonField}}\"`" + `
}

func (s *{{.ClassName}}GetReq) GetId() interface{} { return s.{{.PkGoField}} }

type {{.ClassName}}DeleteReq struct {
	Ids []int ` + "`json:\"ids\"`" + `
}

func (s *{{.ClassName}}DeleteReq) GetId() interface{} { return s.Ids }
`
	legacyGenServiceTemplate = `package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/{{.PackageName}}/models"
	"go-admin/app/{{.PackageName}}/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

type {{.ClassName}} struct {
	service.Service
}

func (e *{{.ClassName}}) GetPage(c *dto.{{.ClassName}}GetPageReq, p *actions.DataPermission, list *[]models.{{.ClassName}}, count *int64) error {
	var data models.{{.ClassName}}
	err := e.Orm.Model(&data).
		Scopes(cDto.MakeCondition(c.GetNeedSearch()), cDto.Paginate(c.GetPageSize(), c.GetPageIndex()), actions.Permission(data.TableName(), p)).
		Find(list).Limit(-1).Offset(-1).Count(count).Error
	if err != nil {
		e.Log.Errorf("{{.ClassName}}Service GetPage error:%s \r\n", err)
		return err
	}
	return nil
}

func (e *{{.ClassName}}) Get(d *dto.{{.ClassName}}GetReq, p *actions.DataPermission, model *models.{{.ClassName}}) error {
	var data models.{{.ClassName}}
	err := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).First(model, d.GetId()).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("查看对象不存在或无权查看")
	}
	return err
}

func (e *{{.ClassName}}) Insert(c *dto.{{.ClassName}}InsertReq) error {
	var data models.{{.ClassName}}
	c.Generate(&data)
	return e.Orm.Create(&data).Error
}

func (e *{{.ClassName}}) Update(c *dto.{{.ClassName}}UpdateReq, p *actions.DataPermission) error {
	var data models.{{.ClassName}}
	e.Orm.Scopes(actions.Permission(data.TableName(), p)).First(&data, c.GetId())
	c.Generate(&data)
	db := e.Orm.Save(&data)
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

func (e *{{.ClassName}}) Remove(d *dto.{{.ClassName}}DeleteReq, p *actions.DataPermission) error {
	var data models.{{.ClassName}}
	db := e.Orm.Model(&data).Scopes(actions.Permission(data.TableName(), p)).Delete(&data, d.GetId())
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		return errors.New("无权删除该数据")
	}
	return nil
}
`
	legacyGenAPIMigrateTemplate = `package version

import (
	"runtime"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/pkg"
	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _{{.GenerateTime}}Test)
}

func _{{.GenerateTime}}Test(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		timeNow := pkg.GetCurrentTime()
		_ = timeNow
		_ = time.Second
		return nil
	})
}
`
)

type legacyGenTableContext struct {
	TableID         int64
	TBName          string
	MLTBName        string
	TableComment    string
	ClassName       string
	TplCategory     string
	PackageName     string
	ModuleName      string
	ModuleFrontName string
	BusinessName    string
	FunctionName    string
	FunctionAuthor  string
	PkColumn        string
	PkGoField       string
	PkJsonField     string
	PkGoType        string
	PkColumnComment string
	Options         string
	TreeCode        string
	TreeParentCode  string
	TreeName        string
	Columns         []legacyGenColumnContext
	ModelColumns    []legacyGenColumnContext
	QueryColumns    []legacyGenColumnContext
	ListColumns     []legacyGenColumnContext
	InsertColumns   []legacyGenColumnContext
	UpdateColumns   []legacyGenColumnContext
	HasTime         bool
	GenerateTime    string
}

type legacyGenColumnContext struct {
	ColumnID      int64
	TableID       int64
	ColumnName    string
	ColumnComment string
	ColumnType    string
	GoType        string
	GoField       string
	JsonField     string
	IsPk          string
	IsIncrement   string
	IsRequired    string
	IsInsert      string
	IsEdit        string
	IsList        string
	IsQuery       string
	QueryType     string
	HtmlType      string
	DictType      string
	Sort          int64
	List          string
	Pk            bool
	Required      bool
	SuperColumn   bool
	UsableColumn  bool
	Increment     bool
	Insert        bool
	Edit          bool
	Query         bool
	Remark        string
	FkTableName   string
	FkLabelID     string
	FkLabelName   string
	Comment       string
	SearchType    string
}

// LegacyGenPreview renders old GoAdmin generator preview templates.
func (s *serviceImpl) LegacyGenPreview(ctx context.Context, tableID int64) (Record, error) {
	genCtx, err := s.legacyGenTable(ctx, tableID, false)
	if err != nil {
		return nil, err
	}
	return s.legacyGenPreview(genCtx)
}

// LegacyGenToProject renders old generated files into a configured output root.
func (s *serviceImpl) LegacyGenToProject(ctx context.Context, tableID int64) (Record, error) {
	genCtx, err := s.legacyGenTable(ctx, tableID, false)
	if err != nil {
		return nil, err
	}
	preview, err := s.legacyGenPreview(genCtx)
	if err != nil {
		return nil, err
	}
	root, err := s.legacyGenRoot(ctx, configKeyLegacyGenRoot, defaultLegacyGenRoot)
	if err != nil {
		return nil, err
	}
	frontRoot, err := s.legacyGenRoot(ctx, configKeyLegacyGenFrontRoot, defaultLegacyGenFrontRoot)
	if err != nil {
		return nil, err
	}
	files := map[string]string{
		filepath.Join(root, "app", genCtx.PackageName, "models", genCtx.TBName+".go"):         gconv.String(preview["template/model.go.template"]),
		filepath.Join(root, "app", genCtx.PackageName, "apis", genCtx.TBName+".go"):           gconv.String(preview["template/api.go.template"]),
		filepath.Join(root, "app", genCtx.PackageName, "router", genCtx.TBName+".go"):         gconv.String(preview["template/router.go.template"]),
		filepath.Join(root, "app", genCtx.PackageName, "service", "dto", genCtx.TBName+".go"): gconv.String(preview["template/dto.go.template"]),
		filepath.Join(root, "app", genCtx.PackageName, "service", genCtx.TBName+".go"):        gconv.String(preview["template/service.go.template"]),
		filepath.Join(frontRoot, "api", genCtx.PackageName, genCtx.MLTBName+".js"):            gconv.String(preview["template/js.go.template"]),
		filepath.Join(frontRoot, "views", genCtx.PackageName, genCtx.MLTBName, "index.vue"):   gconv.String(preview["template/vue.go.template"]),
	}
	paths, err := writeLegacyGenFiles(files)
	if err != nil {
		return nil, err
	}
	return Record{"paths": paths, "root": root, "frontRoot": frontRoot}, nil
}

// LegacyGenAPIToFile renders the old migration API file into a configured root.
func (s *serviceImpl) LegacyGenAPIToFile(ctx context.Context, tableID int64) (Record, error) {
	genCtx, err := s.legacyGenTable(ctx, tableID, false)
	if err != nil {
		return nil, err
	}
	genCtx.GenerateTime = strconv.FormatInt(time.Now().UnixMilli(), 10)
	content, err := renderLegacyGenTemplate("api_migrate", legacyGenAPIMigrateTemplate, genCtx)
	if err != nil {
		return nil, err
	}
	root, err := s.legacyGenRoot(ctx, configKeyLegacyGenRoot, defaultLegacyGenRoot)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, "cmd", "migrate", "migration", "version-local", genCtx.GenerateTime+"_migrate.go")
	paths, err := writeLegacyGenFiles(map[string]string{path: content})
	if err != nil {
		return nil, err
	}
	return Record{"paths": paths, "path": paths[0]}, nil
}

// LegacyGenToDB creates old menu/API permission rows for one generated table.
func (s *serviceImpl) LegacyGenToDB(ctx context.Context, tableID int64) (Record, error) {
	genCtx, err := s.legacyGenTable(ctx, tableID, true)
	if err != nil {
		return nil, err
	}
	actorID := s.actorID(ctx)
	var created []int64
	err = dao.SysMenu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := legacyAlignSysMenuIdentity(ctx, tx); err != nil {
			return err
		}
		insertMenu := func(data do.SysMenu) (int64, error) {
			data.CreateBy = actorID
			data.UpdateBy = actorID
			id, err := tx.Model(dao.SysMenu.Table()).Safe().Ctx(ctx).Data(data).InsertAndGetId()
			if err != nil {
				return 0, err
			}
			created = append(created, id)
			return id, nil
		}
		mainID, err := insertMenu(legacyGenMenuData(genCtx.TBName+"Manage", genCtx.TableComment, "pass", "/"+genCtx.MLTBName, "M", "无", "", 0, "Layout", "0", "0"))
		if err != nil {
			return err
		}
		childID, err := insertMenu(legacyGenMenuData(genCtx.ClassName+"Manage", genCtx.TableComment, "pass", "/"+genCtx.PackageName+"/"+genCtx.MLTBName, "C", "无", genCtx.PackageName+":"+genCtx.BusinessName+":list", mainID, "/"+genCtx.PackageName+"/"+genCtx.MLTBName+"/index", "0", "0"))
		if err != nil {
			return err
		}
		for _, item := range []struct {
			title      string
			permission string
		}{
			{"分页获取" + genCtx.TableComment, genCtx.PackageName + ":" + genCtx.BusinessName + ":query"},
			{"创建" + genCtx.TableComment, genCtx.PackageName + ":" + genCtx.BusinessName + ":add"},
			{"修改" + genCtx.TableComment, genCtx.PackageName + ":" + genCtx.BusinessName + ":edit"},
			{"删除" + genCtx.TableComment, genCtx.PackageName + ":" + genCtx.BusinessName + ":remove"},
		} {
			if _, err := insertMenu(legacyGenMenuData("", item.title, "", genCtx.TBName, "F", "无", item.permission, childID, "", "0", "0")); err != nil {
				return err
			}
		}
		const interfaceID = 63
		apiID, err := insertMenu(legacyGenMenuData(genCtx.TBName, genCtx.TableComment, "bug", genCtx.TBName, "M", "无", "", interfaceID, "", "1", "0"))
		if err != nil {
			return err
		}
		for _, item := range []struct {
			title  string
			path   string
			action string
		}{
			{"分页获取" + genCtx.TableComment, "/api/v1/" + genCtx.ModuleName, "GET"},
			{"根据id获取" + genCtx.TableComment, "/api/v1/" + genCtx.ModuleName + "/:id", "GET"},
			{"创建" + genCtx.TableComment, "/api/v1/" + genCtx.ModuleName, "POST"},
			{"修改" + genCtx.TableComment, "/api/v1/" + genCtx.ModuleName + "/:id", "PUT"},
			{"删除" + genCtx.TableComment, "/api/v1/" + genCtx.ModuleName, "DELETE"},
		} {
			if _, err := insertMenu(legacyGenMenuData("", item.title, "bug", item.path, "A", item.action, "", apiID, "", "1", "0")); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return Record{"menuIds": created, "count": len(created)}, nil
}

func legacyAlignSysMenuIdentity(ctx context.Context, tx gdb.TX) error {
	table := dao.SysMenu.Table()
	sql := fmt.Sprintf(
		"SELECT setval(pg_get_serial_sequence('%s', 'menu_id'), COALESCE((SELECT MAX(menu_id) FROM %s), 0) + 1, false)",
		table,
		table,
	)
	_, err := tx.Exec(sql)
	return err
}

func (s *serviceImpl) legacyGenPreview(genCtx *legacyGenTableContext) (Record, error) {
	templates := map[string]string{
		"template/model.go.template":   legacyGenModelTemplate,
		"template/api.go.template":     legacyGenAPITemplate,
		"template/js.go.template":      legacyGenJSTemplate,
		"template/vue.go.template":     legacyGenVueTemplate,
		"template/router.go.template":  legacyGenRouterTemplate,
		"template/dto.go.template":     legacyGenDTOTemplate,
		"template/service.go.template": legacyGenServiceTemplate,
	}
	result := make(Record, len(templates))
	for key, source := range templates {
		rendered, err := renderLegacyGenTemplate(key, source, genCtx)
		if err != nil {
			return nil, err
		}
		result[key] = rendered
	}
	return result, nil
}

func (s *serviceImpl) legacyGenTable(ctx context.Context, tableID int64, excludeBaseColumns bool) (*legacyGenTableContext, error) {
	if tableID <= 0 {
		return nil, bizerr.NewCode(CodeLegacyGenInvalid)
	}
	table, err := s.GetLegacySystemResource(ctx, "sys-tables", tableID)
	if err != nil {
		return nil, err
	}
	columns, err := s.legacyColumnsByTableIDs(ctx, []int64{tableID})
	if err != nil {
		return nil, err
	}
	genCtx := legacyGenTableFromRecord(table, columns[tableID], excludeBaseColumns)
	if strings.TrimSpace(genCtx.TBName) == "" || strings.TrimSpace(genCtx.ClassName) == "" {
		return nil, bizerr.NewCode(CodeLegacyGenInvalid)
	}
	return genCtx, nil
}

func legacyGenTableFromRecord(table Record, columns []Record, excludeBaseColumns bool) *legacyGenTableContext {
	tbName := strings.TrimSpace(gconv.String(table["tableName"]))
	moduleName := fallbackString(table["moduleName"], tbName)
	moduleFrontName := fallbackString(table["moduleFrontName"], moduleName)
	businessName := fallbackString(table["businessName"], legacyCamelLower(moduleName))
	className := fallbackString(table["className"], legacyPascal(moduleName))
	pkColumn := fallbackString(table["pkColumn"], "id")
	ctx := &legacyGenTableContext{
		TableID:         gconv.Int64(table["tableId"]),
		TBName:          tbName,
		MLTBName:        strings.ReplaceAll(tbName, "_", "-"),
		TableComment:    fallbackString(table["tableComment"], tbName),
		ClassName:       className,
		TplCategory:     strings.TrimSpace(gconv.String(table["tplCategory"])),
		PackageName:     fallbackString(table["packageName"], "admin"),
		ModuleName:      moduleName,
		ModuleFrontName: moduleFrontName,
		BusinessName:    businessName,
		FunctionName:    strings.TrimSpace(gconv.String(table["functionName"])),
		FunctionAuthor:  strings.TrimSpace(gconv.String(table["functionAuthor"])),
		PkColumn:        pkColumn,
		PkGoField:       fallbackString(table["pkGoField"], legacyPascal(pkColumn)),
		PkJsonField:     fallbackString(table["pkJsonField"], legacyCamelLower(pkColumn)),
		Options:         strings.TrimSpace(gconv.String(table["options"])),
		TreeCode:        strings.TrimSpace(gconv.String(table["treeCode"])),
		TreeParentCode:  strings.TrimSpace(gconv.String(table["treeParentCode"])),
		TreeName:        strings.TrimSpace(gconv.String(table["treeName"])),
		PkGoType:        "int",
		PkColumnComment: "id",
		GenerateTime:    strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
	for _, column := range columns {
		item := legacyGenColumnFromRecord(column)
		if item.ColumnName == "" {
			continue
		}
		if item.Pk || item.ColumnName == pkColumn {
			ctx.PkColumn = item.ColumnName
			ctx.PkGoField = item.GoField
			ctx.PkJsonField = item.JsonField
			ctx.PkGoType = item.GoType
			ctx.PkColumnComment = item.ColumnComment
		}
		ctx.Columns = append(ctx.Columns, item)
		if item.GoType == "time.Time" {
			ctx.HasTime = true
		}
		if excludeBaseColumns && legacyGenIsBaseColumn(item.ColumnName) {
			continue
		}
		if !item.Pk && !legacyGenIsAuditField(item.GoField) {
			ctx.ModelColumns = append(ctx.ModelColumns, item)
		}
		if legacyGenTruthy(item.IsQuery) || item.Query {
			ctx.QueryColumns = append(ctx.QueryColumns, item)
		}
		if legacyGenTruthy(item.IsList) || legacyGenTruthy(item.List) {
			ctx.ListColumns = append(ctx.ListColumns, item)
		}
		if !item.Pk && !legacyGenIsAuditField(item.GoField) && (legacyGenTruthy(item.IsInsert) || item.Insert) {
			ctx.InsertColumns = append(ctx.InsertColumns, item)
		}
		if !item.Pk && !legacyGenIsAuditField(item.GoField) && (legacyGenTruthy(item.IsEdit) || item.Edit) {
			ctx.UpdateColumns = append(ctx.UpdateColumns, item)
		}
	}
	if ctx.PkGoField == "" {
		ctx.PkGoField = "Id"
	}
	if ctx.PkJsonField == "" {
		ctx.PkJsonField = "id"
	}
	return ctx
}

func legacyGenColumnFromRecord(record Record) legacyGenColumnContext {
	columnName := strings.TrimSpace(gconv.String(record["columnName"]))
	goType := fallbackString(record["goType"], "string")
	goField := fallbackString(record["goField"], legacyPascal(columnName))
	jsonField := fallbackString(record["jsonField"], legacyCamelLower(columnName))
	columnComment := strings.TrimSpace(gconv.String(record["columnComment"]))
	if columnComment == "" {
		columnComment = goField
	}
	item := legacyGenColumnContext{
		ColumnID:      gconv.Int64(record["columnId"]),
		TableID:       gconv.Int64(record["tableId"]),
		ColumnName:    columnName,
		ColumnComment: columnComment,
		ColumnType:    fallbackString(record["columnType"], "varchar(255)"),
		GoType:        goType,
		GoField:       goField,
		JsonField:     jsonField,
		IsPk:          strings.TrimSpace(gconv.String(record["isPk"])),
		IsIncrement:   strings.TrimSpace(gconv.String(record["isIncrement"])),
		IsRequired:    strings.TrimSpace(gconv.String(record["isRequired"])),
		IsInsert:      strings.TrimSpace(gconv.String(record["isInsert"])),
		IsEdit:        strings.TrimSpace(gconv.String(record["isEdit"])),
		IsList:        strings.TrimSpace(gconv.String(record["isList"])),
		IsQuery:       strings.TrimSpace(gconv.String(record["isQuery"])),
		QueryType:     strings.TrimSpace(gconv.String(record["queryType"])),
		HtmlType:      strings.TrimSpace(gconv.String(record["htmlType"])),
		DictType:      strings.TrimSpace(gconv.String(record["dictType"])),
		Sort:          gconv.Int64(record["sort"]),
		List:          strings.TrimSpace(gconv.String(record["list"])),
		Pk:            gconv.Bool(record["pk"]) || strings.EqualFold(gconv.String(record["isPk"]), "1"),
		Required:      gconv.Bool(record["required"]) || strings.EqualFold(gconv.String(record["isRequired"]), "1"),
		SuperColumn:   gconv.Bool(record["superColumn"]),
		UsableColumn:  gconv.Bool(record["usableColumn"]),
		Increment:     gconv.Bool(record["increment"]) || strings.EqualFold(gconv.String(record["isIncrement"]), "1"),
		Insert:        gconv.Bool(record["insert"]) || strings.EqualFold(gconv.String(record["isInsert"]), "1"),
		Edit:          gconv.Bool(record["edit"]) || strings.EqualFold(gconv.String(record["isEdit"]), "1"),
		Query:         gconv.Bool(record["query"]) || strings.EqualFold(gconv.String(record["isQuery"]), "1"),
		Remark:        strings.TrimSpace(gconv.String(record["remark"])),
		FkTableName:   strings.TrimSpace(gconv.String(record["fkTableName"])),
		FkLabelID:     strings.TrimSpace(gconv.String(record["fkLabelId"])),
		FkLabelName:   strings.TrimSpace(gconv.String(record["fkLabelName"])),
		Comment:       columnComment,
	}
	item.SearchType = legacyGenSearchType(item.QueryType)
	return item
}

func renderLegacyGenTemplate(name string, source string, data any) (string, error) {
	tmpl, err := template.New(name).Parse(source)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeLegacyGenFailed)
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", bizerr.WrapCode(err, CodeLegacyGenFailed)
	}
	return buffer.String(), nil
}

func writeLegacyGenFiles(files map[string]string) ([]string, error) {
	paths := make([]string, 0, len(files))
	for path, content := range files {
		cleanPath := filepath.Clean(path)
		if strings.TrimSpace(cleanPath) == "" {
			return nil, bizerr.NewCode(CodeLegacyGenInvalid)
		}
		if err := gfile.Mkdir(filepath.Dir(cleanPath)); err != nil {
			return nil, bizerr.WrapCode(err, CodeLegacyGenFailed)
		}
		if err := os.WriteFile(cleanPath, []byte(content), 0o644); err != nil {
			return nil, bizerr.WrapCode(err, CodeLegacyGenFailed)
		}
		paths = append(paths, cleanPath)
	}
	return paths, nil
}

func (s *serviceImpl) legacyGenRoot(ctx context.Context, key string, defaultValue string) (string, error) {
	root, err := s.configSvc.String(ctx, key, defaultValue)
	if err != nil {
		return "", err
	}
	root = strings.TrimSpace(root)
	if root == "" {
		root = defaultValue
	}
	return filepath.Clean(root), nil
}

func legacyGenMenuData(menuName string, title string, icon string, path string, menuType string, action string, permission string, parentID int64, component string, visible string, isFrame string) do.SysMenu {
	return do.SysMenu{
		MenuName:   menuName,
		Title:      title,
		Icon:       icon,
		Path:       path,
		Paths:      path,
		MenuType:   menuType,
		Action:     action,
		Permission: permission,
		ParentId:   parentID,
		NoCache:    false,
		Breadcrumb: "",
		Component:  component,
		Sort:       0,
		Visible:    visible,
		IsFrame:    isFrame,
	}
}

func legacyGenSearchType(queryType string) string {
	switch strings.ToUpper(strings.TrimSpace(queryType)) {
	case "EQ":
		return "exact"
	case "NE":
		return "iexact"
	case "LIKE":
		return "contains"
	case "GT":
		return "gt"
	case "GTE":
		return "gte"
	case "LT":
		return "lt"
	case "LTE":
		return "lte"
	default:
		return "exact"
	}
}

func legacyGenTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func legacyGenIsAuditField(goField string) bool {
	switch goField {
	case "CreatedAt", "UpdatedAt", "DeletedAt", "CreateBy", "UpdateBy":
		return true
	default:
		return false
	}
}

func legacyGenIsBaseColumn(column string) bool {
	switch strings.ToLower(strings.TrimSpace(column)) {
	case "id", "create_by", "update_by", "created_at", "updated_at", "deleted_at":
		return true
	default:
		return false
	}
}

func legacyPascal(value string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		builder.WriteString(strings.ToUpper(lower[:1]))
		if len(lower) > 1 {
			builder.WriteString(lower[1:])
		}
	}
	return builder.String()
}

func legacyCamelLower(value string) string {
	pascal := legacyPascal(value)
	if pascal == "" {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}
